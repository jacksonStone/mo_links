package http_server

import (
	"fmt"
	"mo_links/models"
	"net/http"
	"net/url"
	"time"
)

func initializeEmailTokenRoutes() {
	http.HandleFunc("/____reserved/api/verify_email", signupEmailVerificationHandler)
}

func signupEmailVerificationHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	user, err := getUserInCookies(r)
	if err != nil {
		fmt.Println("Error getting user in cookies", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fullUser, err := models.GetUserById(user.Id)
	if err != nil {
		fmt.Println("Error getting user by id", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if fullUser.VerifiedEmail {
		// Already Verified - just leave
		fmt.Println("Already Verified")
		w.WriteHeader(http.StatusOK)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	encodedToken, err := url.QueryUnescape(fullUser.VerificationToken)
	if err != nil {
		fmt.Println("Error unescaping verification token", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if encodedToken != token {
		fmt.Println("Verification token mismatch: ", encodedToken, token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if fullUser.VerificationTokenExpiresAt.Before(time.Now()) {
		fmt.Println("Verification token expired")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fullUser.VerifiedEmail = true
	models.SetEmailToVerified(user.Id)
	refreshCookie(fullUser, w)
	// redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}
