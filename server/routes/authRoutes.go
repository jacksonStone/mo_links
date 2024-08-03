package routes

import (
	"encoding/json"
	"fmt"
	"mo_links/auth"
	"mo_links/db"
	"net/http"
	"strconv"
	"time"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func InitAuthRoutes() {
	http.HandleFunc("/____reserved/api/login", loginEndpoint)
	http.HandleFunc("/____reserved/api/signup", signupEndpoint)
	http.HandleFunc("/____reserved/api/test_cookie", testCookieEndpoint)
}

func testCookieEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
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
	var request loginRequest
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
