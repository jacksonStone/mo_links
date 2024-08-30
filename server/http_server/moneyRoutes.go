package http_server

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	ItemId string `json:"itemId"`
}
type fatStacksCheckoutRequest struct {
	AppName            string   `json:"appName"` // Should be the domain of the app (Ex: "www.fatstacks.io")
	UserId             string   `json:"userId"`
	UserEmail          string   `json:"userEmail"`
	Items              []string `json:"items"` // Should be a list of product Ids as specified in the product_catalog.json for the app
	FulfillmentUrl     string   `json:"fulfillmentUrl"`
	UndoFulfillmentUrl string   `json:"undoFulfillmentUrl"`
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
	var fatStacksCheckoutRequest fatStacksCheckoutRequest
	fatStacksCheckoutRequest.AppName = siteDomain
	fatStacksCheckoutRequest.UserId = strconv.FormatInt(user.Id, 10)
	fatStacksCheckoutRequest.UserEmail = user.Email
	fatStacksCheckoutRequest.Items = []string{checkoutBody.ItemId}
	fatStacksCheckoutRequest.FulfillmentUrl = "https://" + siteDomain + "/____reserved/fulfill_checkout"
	fatStacksCheckoutRequest.UndoFulfillmentUrl = "https://" + siteDomain + "/____reserved/unfulfill_checkout"

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
		fmt.Println("getCheckoutEndpoint - cant decode")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(fatStacksCheckoutResponse)

}

type fatStacksCheckoutResponse struct {
	RedirectUrl string `json:"redirectUrl"`
}

func getFulfillCheckoutEndpoint(w http.ResponseWriter, r *http.Request) {

}

func getUnfulfillCheckoutEndpoint(w http.ResponseWriter, r *http.Request) {

}
