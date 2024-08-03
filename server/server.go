package main

// TODO Leverage organization endpoints (Still need to be able to add people)

import (
	"encoding/json"
	"fmt"
	"log"
	"mo_links/auth"
	"mo_links/common"
	"mo_links/db"
	"mo_links/models"
	"mo_links/routes"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type AddLinkRequest struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type CreateOrganizationRequest struct {
	Name string `json:"name"`
}
type AssignMemberRequest struct {
	UserId         int64  `json:"userId"`
	OrganizationId int64  `json:"organizationId"`
	Role           string `json:"role"`
}
type GetOrganizationMembersRequest struct {
	OrganizationId int64 `json:"organizationId"`
}
type UpdateActiveOrganizationRequest struct {
	OrganizationId int64 `json:"organizationId"`
}

func main() {
	db.InitializeDB()
	auth.InitAuth()
	routes.InitializeInvitesRoute()
	http.HandleFunc("/____reserved/api/login", loginEndpoint)
	http.HandleFunc("/____reserved/api/signup", signupEndpoint)
	http.HandleFunc("/____reserved/api/test_cookie", testCookieEndpoint)
	http.HandleFunc("/____reserved/api/add", addLinkEndpoint)
	http.HandleFunc("/____reserved/api/me", meEndpoint)
	http.HandleFunc("/____reserved/api/organizations", organizationsEndpoint)
	http.HandleFunc("/____reserved/api/organization/make_active", makeActiveOrganizationEndpoint)
	http.HandleFunc("/____reserved/api/organization/create", createOrganizationEndpoint)
	http.HandleFunc("/____reserved/api/organization/assign_member", assignMemberEndpoint)
	http.HandleFunc("/____reserved/api/organization/members", getOrganizationMembersEndpoint)

	routes.InitStaticRoutes()

	http.HandleFunc("/", handleAttemptedMoLink)

	err := http.ListenAndServe(":3003", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func meEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userDetails, err := models.GetUserDetails(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDetails)
}
func decodeLink(r *http.Request, organizationId int64) ([]string, error) {

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		return []string{}, nil
	}
	links, err := models.GetMatchingLinks(organizationId, path)
	fmt.Println(links)
	if err != nil {
		return []string{}, err
	}
	return links, nil
}

func addLinkEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// Allow CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	// Handle preflight request
	if r.Method == http.MethodOptions {
		return
	}
	var request AddLinkRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = models.AddLink(request.Url, request.Name, user.Id, user.ActiveOrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusExpectationFailed)
		return
	}
}

func testCookieEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Write([]byte(fmt.Sprintf("Hello, %s", user.Email)))
}

func handleAttemptedMoLink(w http.ResponseWriter, r *http.Request) {
	// For when running locally without the reverse proxy
	if strings.HasSuffix(r.URL.Path, "/____reserved/_ping") {
		w.Write([]byte("OK"))
		return
	}
	if r.URL.Path == "/" {
		routes.ServeHomePage(w)
		return
	}
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.Redirect(w, r, "/____reserved/login_page?next="+url.QueryEscape(r.URL.Path), http.StatusFound)
		return
	}
	links, err := decodeLink(r, user.ActiveOrganizationId)
	if err != nil {
		// TODO improve error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(links) == 0 {
		// TODO Open up the home page to create a new link
		routes.ServeHomePage(w)
		return
	}
	link := links[0]
	// Redirect user to link, but don't cache the result
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.Redirect(w, r, link, http.StatusFound)
}
func signupEndpoint(w http.ResponseWriter, r *http.Request) {
	// Grab email and password from request
	// Verify the user does not exist with that email, create user, then login.
	var request LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, _ := db.DbGetUserByEmail(request.Email)
	if userId != 0 {
		http.Error(w, "User with that email already exists", http.StatusBadRequest)
		return
	}
	salt := auth.Auth.GenerateSalt()
	verificationToken := auth.Auth.GenerateSalt() // Get a different random string for the verification token so that the "actual" salt will not be sent over email
	hashedPassword := auth.Auth.GetHashedPasswordFromRawTextPassword(request.Password, salt)
	verificationExpiration := int64(time.Now().Add(7 * 24 * time.Hour).Unix())
	// Create the user
	err = db.DbSignupUser(request.Email, hashedPassword, salt, verificationToken, verificationExpiration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	attemptLogin(w, request.Email, request.Password)
}
func loginEndpoint(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	attemptLogin(w, request.Email, request.Password)
}

func attemptLogin(w http.ResponseWriter, userEmail string, plainTextPassword string) {
	userId, err := db.DbGetUserByEmail(userEmail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	attemptedCookie, err := auth.Auth.AttemptLoginAndGetCookie(strconv.Itoa(int(userId)), plainTextPassword)
	if err != nil {
		http.Error(w, "Invalid Login", http.StatusBadRequest)
		return
	}
	// Write the cookie
	w.Header().Set("Set-Cookie", attemptedCookie)
	w.Write([]byte("OK"))
}

func refreshCookie(user common.User, w http.ResponseWriter) {
	cookie, err := auth.Auth.CreateNewCookie(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Set-Cookie", cookie)
}

func organizationsEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	organizations, err := models.GetMatchingOrganizations(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organizations)
}

func createOrganizationEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request CreateOrganizationRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = models.CreateOrganization(request.Name, user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func assignMemberEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request AssignMemberRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	role, err := models.GetUserRoleInOrganization(user.Id, request.OrganizationId)

	if err != nil {
		http.Error(w, "Error checking user permissions", http.StatusInternalServerError)
		return
	}
	if role == "" {
		http.Error(w, "User is not a member of this organization", http.StatusForbidden)
		return
	}

	org, _ := models.GetOrganizationById(request.OrganizationId)
	if org.IsPersonal {
		http.Error(w, "Cannot assign members to personal organization", http.StatusBadRequest)
		return
	}

	if !models.RoleCanAddRole(role, request.Role) {
		http.Error(w, "Unauthorized to assign "+request.Role+"s to this organization", http.StatusForbidden)
		return
	}

	err = models.AssignMemberToOrganization(request.UserId, request.Role, request.OrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func makeActiveOrganizationEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var request UpdateActiveOrganizationRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if request.OrganizationId == 0 {
		http.Error(w, "must provide target organization ID", http.StatusBadRequest)
		return
	}
	organizations, err := models.GetMatchingOrganizations(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If we are in that org, we can assign it as our active org
	for _, org := range organizations {
		if org.Id == request.OrganizationId {
			err = db.DbSetUserActiveOrganization(user.Id, request.OrganizationId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Refresh login cookie
			fullUser, err := auth.Auth.GetUser(strconv.Itoa(int(user.Id)))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			refreshCookie(fullUser, w)
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "user is not a member of that organization", http.StatusForbidden)
}

func getOrganizationMembersEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request GetOrganizationMembersRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	role, err := models.GetUserRoleInOrganization(user.Id, request.OrganizationId)

	if err != nil {
		http.Error(w, "Error checking user permissions", http.StatusInternalServerError)
		return
	}
	if role == "" {
		http.Error(w, "User is not a member of this organization", http.StatusForbidden)
		return
	}

	if !models.RoleCanViewMembers(role) {
		http.Error(w, "User is not authorized to view members of this organization", http.StatusForbidden)
		return
	}

	members, err := models.GetOrganizationMembers(request.OrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
