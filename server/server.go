package main

// TODO Leverage organization endpoints
// TODO Configure mo links match to only match against active organization for the user, maybe package that into the cookie?
//     SUB TODO: Make sure that the user can configure a org in focus
// TODO Create an easy UI to swap between organizations

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//go:embed static
var static embed.FS

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

func main() {
	initializeDB()
	InitAuth()

	http.HandleFunc("/____reserved/api/login", loginEndpoint)
	http.HandleFunc("/____reserved/api/signup", signupEndpoint)
	http.HandleFunc("/____reserved/api/test-cookie", testCookieEndpoint)
	http.HandleFunc("/____reserved/api/add", addLinkEndpoint)
	http.HandleFunc("/____reserved/api/organizations", organizationsEndpoint)
	http.HandleFunc("/____reserved/api/organization/create", createOrganizationEndpoint)
	http.HandleFunc("/____reserved/api/organization/assign-member", assignMemberEndpoint)
	http.HandleFunc("/____reserved/api/organization/members", getOrganizationMembersEndpoint)

	http.HandleFunc("/____reserved/privacy_policy", getPrivacyPolicyEndpoint)
	http.HandleFunc("/____reserved/login_page", loginPageEndpoint)

	http.HandleFunc("/favicon.ico", faviconEndpoint)

	http.HandleFunc("/", handleAttemptedMoLink)

	http.ListenAndServe(":3003", nil)
}

func returnStaticFile(w http.ResponseWriter, path string) {
	bytes, err := static.ReadFile(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func faviconEndpoint(w http.ResponseWriter, r *http.Request) {
	// Set appropriate content type
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	returnStaticFile(w, "static/logo-16.png")
}
func serveHomePage(w http.ResponseWriter) {
	returnStaticFile(w, "static/index.html")
}

func getPrivacyPolicyEndpoint(w http.ResponseWriter, r *http.Request) {
	returnStaticFile(w, "static/privacy_policy.html")
}

func loginPageEndpoint(w http.ResponseWriter, r *http.Request) {
	returnStaticFile(w, "static/login.html")
}

func decodeLink(r *http.Request, organizationId int64) ([]string, error) {

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		return []string{}, nil
	}
	links, err := getMatchingLinks(organizationId, path)
	fmt.Println(links)
	if err != nil {
		return []string{}, err
	}
	return links, nil
}

func getAuthenticatedUser(r *http.Request) (TrimmedUser, error) {
	// Log all cookies
	var authCookie string
	for _, cookie := range r.Cookies() {
		if cookie.Name == Auth.CookieName {
			authCookie = cookie.Value
		}
	}
	if authCookie == "" {
		return TrimmedUser{}, errors.New("no auth cookie found")
	}
	// Get raw cookie header
	rawCookieHeader := r.Header.Get("Cookie")
	if rawCookieHeader == "" {
		return TrimmedUser{}, errors.New("no raw cookie header found")
	}

	decryptedCookie, err := Auth.AttemptCookieDecryption(rawCookieHeader)
	if err != nil {
		return TrimmedUser{}, err
	}
	// parse cookie as json
	var user TrimmedUser
	err = json.Unmarshal([]byte(decryptedCookie), &user)
	if err != nil {
		return TrimmedUser{}, err
	}
	return user, nil
}

func addLinkEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getAuthenticatedUser(r)
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
	err = addLink(request.Url, request.Name, user.Id, user.ActiveOrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusExpectationFailed)
		return
	}
}

func testCookieEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getAuthenticatedUser(r)
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
	user, err := getAuthenticatedUser(r)
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
		serveHomePage(w)
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
	userId, _ := dbGetUserByEmail(request.Email)
	if userId != 0 {
		http.Error(w, "User with that email already exists", http.StatusBadRequest)
		return
	}
	salt := Auth.GenerateSalt()
	verificationToken := Auth.GenerateSalt() // Get a different random string for the verification token so that the "actual" salt will not be sent over email
	hashedPassword := Auth.GetHashedPasswordFromRawTextPassword(request.Password, salt)
	verificationExpiration := int64(time.Now().Add(7 * 24 * time.Hour).Unix())
	// Create the user
	err = dbSignupUser(request.Email, hashedPassword, salt, verificationToken, verificationExpiration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	attemptLogin(w, request.Email, request.Password)
}
func loginEndpoint(w http.ResponseWriter, r *http.Request) {
	// Grab email and password from request
	// Check if email and password are correct
	// If correct, return cookie
	// If incorrect, return error
	var request LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	attemptLogin(w, request.Email, request.Password)
}

func attemptLogin(w http.ResponseWriter, userEmail string, plainTextPassword string) {
	userId, err := dbGetUserByEmail(userEmail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	attemptedCookie, err := Auth.AttemptLoginAndGetCookie(strconv.Itoa(int(userId)), plainTextPassword)
	if err != nil {
		http.Error(w, "Invalid Login", http.StatusBadRequest)
		return
	}
	// Write the cookie
	w.Header().Set("Set-Cookie", attemptedCookie)
	w.Write([]byte("OK"))
}

func organizationsEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	organizations, err := getMatchingOrganizations(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organizations)
}

func createOrganizationEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getAuthenticatedUser(r)
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

	err = createOrganization(request.Name, user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func assignMemberEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getAuthenticatedUser(r)
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

	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	role, err := getUserRoleInOrganization(user.Id, request.OrganizationId)

	if err != nil {
		http.Error(w, "Error checking user permissions", http.StatusInternalServerError)
		return
	}
	if role == "" {
		http.Error(w, "User is not a member of this organization", http.StatusForbidden)
		return
	}

	if !roleCanAddRole(role, request.Role) {
		http.Error(w, "Unauthorized to assign "+request.Role+"s to this organization", http.StatusForbidden)
		return
	}

	err = assignMemberToOrganization(request.UserId, request.Role, request.OrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getOrganizationMembersEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getAuthenticatedUser(r)
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
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	role, err := getUserRoleInOrganization(user.Id, request.OrganizationId)

	if err != nil {
		http.Error(w, "Error checking user permissions", http.StatusInternalServerError)
		return
	}
	if role == "" {
		http.Error(w, "User is not a member of this organization", http.StatusForbidden)
		return
	}

	if !roleCanViewMembers(role) {
		http.Error(w, "User is not authorized to view members of this organization", http.StatusForbidden)
		return
	}

	members, err := getOrganizationMembers(request.OrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
