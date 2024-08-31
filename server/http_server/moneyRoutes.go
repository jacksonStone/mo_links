package http_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mo_links/models"
	"net/http"
	"os"
	"strconv"
)

const siteDomain = "www.molinks.me"

func initializeMoneyRoutes() {
	http.HandleFunc("/____reserved/checkout", getCheckoutEndpoint)
	http.HandleFunc("/____reserved/product_catalog_image", getProductCatalogImageEndpoint)
	http.HandleFunc("/____reserved/fulfill_checkout", getFulfillCheckoutEndpoint)
	http.HandleFunc("/____reserved/unfulfill_checkout", getUnfulfillCheckoutEndpoint)
}

type CheckoutRequest struct {
	ItemId         string `json:"itemId"`
	OrganizationId string `json:"organizationId"`
}
type fatStacksCheckoutRequest struct {
	AppName            string            `json:"appName"` // Should be the domain of the app (Ex: "www.fatstacks.io")
	UserId             string            `json:"userId"`
	UserEmail          string            `json:"userEmail"`
	Items              []string          `json:"items"` // Should be a list of product Ids as specified in the product_catalog.json for the app
	FulfillmentUrl     string            `json:"fulfillmentUrl"`
	UndoFulfillmentUrl string            `json:"undoFulfillmentUrl"`
	Metadata           map[string]string `json:"metadata"`
}

func getProductCatalogImageEndpoint(w http.ResponseWriter, r *http.Request) {
	// allow cors
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "image/png")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	returnStaticFile(w, "static/logo-128.png")
}

func getCheckoutEndpoint(w http.ResponseWriter, r *http.Request) {
	returnURL := "https://" + siteDomain
	if os.Getenv("NODE_ENV") == "development" {
		returnURL = "http://localhost:3003"
	}
	fmt.Println("getCheckoutEndpoint")
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		fmt.Println(err)
		fmt.Println("getCheckoutEndpoint - cant auth")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	fatStacksUrl := os.Getenv("FAT_STACKS_URL")
	fmt.Println(fatStacksUrl)
	// Decode the request body
	var checkoutBody CheckoutRequest
	err = json.NewDecoder(r.Body).Decode(&checkoutBody)
	if err != nil {
		fmt.Println(err)
		fmt.Println("getCheckoutEndpoint - cant decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if checkoutBody.OrganizationId == "" {
		fmt.Println("getCheckoutEndpoint - organizationId is required")
		http.Error(w, "OrganizationId is required", http.StatusBadRequest)
		return
	}
	if checkoutBody.ItemId == "" {
		fmt.Println("getCheckoutEndpoint - itemId is required")
		http.Error(w, "ItemId is required", http.StatusBadRequest)
		return
	}
	organizationId, err := strconv.ParseInt(checkoutBody.OrganizationId, 10, 64)
	if err != nil {
		fmt.Println("Error parsing OrganizationId:", err)
		http.Error(w, "Invalid OrganizationId", http.StatusBadRequest)
		return
	}
	userRole, err := models.GetUserRoleInOrganization(user.Id, organizationId)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !models.RoleCanBuySubscription(userRole) {
		http.Error(w, "User is not allowed to buy subscription for this organization", http.StatusUnauthorized)
		return
	}
	var fatStacksCheckoutRequest fatStacksCheckoutRequest
	fatStacksCheckoutRequest.AppName = siteDomain
	fatStacksCheckoutRequest.UserId = strconv.FormatInt(user.Id, 10)
	fatStacksCheckoutRequest.UserEmail = user.Email
	fatStacksCheckoutRequest.Items = []string{checkoutBody.ItemId}
	fatStacksCheckoutRequest.FulfillmentUrl = returnURL + "/____reserved/fulfill_checkout"
	fatStacksCheckoutRequest.UndoFulfillmentUrl = returnURL + "/____reserved/unfulfill_checkout"
	fatStacksCheckoutRequest.Metadata = map[string]string{
		//Required for fulfillment
		"organizationId": checkoutBody.OrganizationId,
	}
	fatStacksCheckoutRequestJson, err := json.Marshal(fatStacksCheckoutRequest)
	if err != nil {
		fmt.Println(err)
		fmt.Println("getCheckoutEndpoint - cant marshal")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Make a post request to the fat stacks url
	resp, err := http.Post(fatStacksUrl+"/create_checkout", "application/json", bytes.NewBuffer(fatStacksCheckoutRequestJson))
	if err != nil {
		fmt.Println(err)
		fmt.Println("getCheckoutEndpoint - cant post")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var fatStacksCheckoutResponse fatStacksCheckoutResponse
	err = json.NewDecoder(resp.Body).Decode(&fatStacksCheckoutResponse)
	if err != nil {
		fmt.Println(err)
		fmt.Println("getCheckoutEndpoint - cant decode response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(fatStacksCheckoutResponse)
	// Redirect to the fat stacks checkout url
	http.Redirect(w, r, fatStacksCheckoutResponse.RedirectUrl, http.StatusSeeOther)
}

type fatStacksCheckoutResponse struct {
	RedirectUrl string `json:"redirectUrl"`
}
type checkoutFulfillmentResult struct {
	UserId   string   `json:"userId"`
	Items    []string `json:"items"`
	Metadata string   `json:"metadata"` // JSON Stringified... I hope.
}

func getFulfillCheckoutEndpoint(w http.ResponseWriter, r *http.Request) {
	// Check that the request contains the fat stacks secret
	fatStacksSecret := os.Getenv("FAT_STACKS_SECRET")
	if r.Header.Get("X-Fat-Stacks-Secret") != fatStacksSecret {
		fmt.Println("getFulfillCheckoutEndpoint - invalid secret")
		http.Error(w, "Invalid secret", http.StatusUnauthorized)
		return
	}
	// Decode the request body
	var checkoutFulfillmentResult checkoutFulfillmentResult
	err := json.NewDecoder(r.Body).Decode(&checkoutFulfillmentResult)
	if err != nil {
		fmt.Println(err)
		fmt.Println("getFulfillCheckoutEndpoint - cant decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	metadata := make(map[string]string)
	err = json.Unmarshal([]byte(checkoutFulfillmentResult.Metadata), &metadata)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("getFulfillCheckoutEndpoint - fulfillment result", checkoutFulfillmentResult)
	//We only support one item for now
	product, err := getProductFromProductCatalog(checkoutFulfillmentResult.Items[0])
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	months := 0
	if product.Recurring.Interval == "month" {
		months = 1
	} else if product.Recurring.Interval == "year" {
		months = 12
	}
	// get the organization from the metadata
	organizationId := metadata["organizationId"]
	fmt.Println("getFulfillCheckoutEndpoint - organizationId", organizationId)
	organizationIdInt, err := strconv.ParseInt(organizationId, 10, 64)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	organization, err := models.GetOrganizationById(organizationIdInt)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = models.SetSubscriptionToActive(organization.Id, months)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	organization, err = models.GetOrganizationById(organizationIdInt)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("getFulfillCheckoutEndpoint - organization after set subscription to active", organization)

	// TODO:: Set organization to active_subscription

	fmt.Println("getFulfillCheckoutEndpoint - organization", organization)
	fmt.Println("getFulfillCheckoutEndpoint - months", months)

}

func getUnfulfillCheckoutEndpoint(w http.ResponseWriter, r *http.Request) {
	// Check that the request contains the fat stacks secret
	fatStacksSecret := os.Getenv("FAT_STACKS_SECRET")
	if r.Header.Get("X-Fat-Stacks-Secret") != fatStacksSecret {
		fmt.Println("getUnfulfillCheckoutEndpoint - invalid secret")
		http.Error(w, "Invalid secret", http.StatusUnauthorized)
		return
	}
	// Should not be exposed to the public
	fmt.Println("TODO: handle the unfulfillment")
}

type ProductCatalog map[string]Website

type Website struct {
	StripeKeys        StripeKeys         `json:"stripe_keys"`
	SendgridKeys      SendgridKeys       `json:"sendgrid_keys"`
	DefaultCancelURL  string             `json:"default_cancel_url"`
	DefaultSuccessURL string             `json:"default_success_url"`
	Products          map[string]Product `json:"products"`
}

type StripeKeys struct {
	PublicKey     string `json:"public_key"`
	PrivateKey    string `json:"private_key"`
	WebhookSecret string `json:"webhook_secret"`
}

type SendgridKeys struct {
	APIKey string `json:"api_key"`
}

type Product struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       int        `json:"price"`
	Savings     *int       `json:"savings,omitempty"`
	Image       string     `json:"image"`
	Recurring   *Recurring `json:"recurring,omitempty"`
}

type Recurring struct {
	Interval string `json:"interval"`
}

var productCatalog ProductCatalog

func getProductCatalog() ProductCatalog {
	if productCatalog != nil {
		return productCatalog
	}
	productCatalogBytes, err := static.ReadFile("static/product_catalog.json")
	if err != nil {
		log.Fatalf("Failed to read product catalog: %v", err)
	}

	err = json.Unmarshal(productCatalogBytes, &productCatalog)
	if err != nil {
		log.Fatalf("Failed to unmarshal product catalog: %v", err)
	}
	return productCatalog
}

func getProductFromProductCatalog(productId string) (Product, error) {
	productCatalog := getProductCatalog()
	websiteCatalog, ok := productCatalog[siteDomain]
	if !ok {
		return Product{}, fmt.Errorf("website not found in product catalog")
	}
	product, ok := websiteCatalog.Products[productId]
	if !ok {
		return Product{}, fmt.Errorf("product not found in product catalog")
	}
	return product, nil
}
