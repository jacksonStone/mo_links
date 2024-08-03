package http_server

import (
	"encoding/json"
	"fmt"
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
	user, err := getUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Write([]byte(fmt.Sprintf("Hello, %s", user.Email)))
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
