package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func main() {
	initializeDB()
	InitAuth()
	// cookie, err := Auth.AttemptLoginAndGetCookie("1", "password")
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("cookie:", cookie)
	// 	decryptedCookie, err := Auth.AttemptCookieDecryption(cookie)
	// 	if err != nil {
	// 		fmt.Println("Error decrypting cookie:", err)
	// 	} else {
	// 		fmt.Println("Decrypted Cookie:", decryptedCookie)
	// 	}
	// }
	http.HandleFunc("/____reserved/privacy_policy", getPrivacyPolicyEndpoint)
	http.HandleFunc("/____reserved/api/login", loginEndpoint)
	http.HandleFunc("/____reserved/login_page", loginPageEndpoint)

	http.HandleFunc("/____reserved/api/test-cookie", testCookieEndpoint)

	http.HandleFunc("/____reserved/api/add", addLinkEndpoint)
	http.HandleFunc("/favicon.ico", faviconEndpoint)

	http.HandleFunc("/", handleAttemptedMoLink)

	http.ListenAndServe(":3003", nil)
}

func faviconEndpoint(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/logo-16.png")
}

func decodeLink(r *http.Request, userId int32) ([]string, error) {

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		return []string{}, nil
	}
	links, err := getMatchingLinks(userId, path)
	if err != nil {
		return []string{}, err
	}
	return links, nil
}

func serveHomePage(w http.ResponseWriter) {
	bytes, err := static.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func getPrivacyPolicyEndpoint(w http.ResponseWriter, r *http.Request) {
	bytes, err := static.ReadFile("static/privacy_policy.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
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
	err = addLink(request.Url, request.Name, user.Id)
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
	user, err := getAuthenticatedUser(r)
	if err != nil {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.Redirect(w, r, "/____reserved/login_page?next="+url.QueryEscape(r.URL.Path), http.StatusFound)
		return
	}
	links, err := decodeLink(r, user.Id)
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

func loginPageEndpoint(w http.ResponseWriter, r *http.Request) {
	bytes, err := static.ReadFile("static/login.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
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
	userId, err := dbGetUserByEmail(request.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if userId == 0 {
		http.Error(w, "Invalid Login", http.StatusBadRequest)
		return
	}
	attemptedCookie, err := Auth.AttemptLoginAndGetCookie(strconv.Itoa(int(userId)), request.Password)
	if err != nil {
		http.Error(w, "Invalid Login", http.StatusBadRequest)
		return
	}
	// Write the cookie
	w.Header().Set("Set-Cookie", attemptedCookie)
	w.Write([]byte("OK"))
}
