package http_server

import (
	"encoding/json"
	"mo_links/db"
	"mo_links/models"
	"net/http"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func initializeAuthRoutes() {
	http.HandleFunc("/____reserved/api/login", loginEndpoint)
	http.HandleFunc("/____reserved/api/signup", signupEndpoint)
	http.HandleFunc("/____reserved/api/test_cookie", testCookieEndpoint)
}

func testCookieEndpoint(w http.ResponseWriter, r *http.Request) {
	// Allow CORS - the extension needs this
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	// Handle preflight request
	if r.Method == http.MethodOptions {
		return
	}
	// mo/goo
	user, err := getUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// Send user as json
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func signupEndpoint(w http.ResponseWriter, r *http.Request) {
	// Grab email and password from request
	// Verify the user does not exist with that email, create user, then login.
	var request loginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = models.SignupUser(request.Email, request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := models.GetUserIdByEmail(request.Email)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	attemptLogin(w, userId, request.Password)
}

func loginEndpoint(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := db.DbGetUserByEmail(request.Email)
	if err != nil {
		http.Error(w, "invalid login, no user with that email or password", http.StatusBadRequest)
		return
	}
	attemptLogin(w, userId, request.Password)
}
