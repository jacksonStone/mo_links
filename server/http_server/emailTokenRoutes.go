package http_server

import (
	"fmt"
	"mo_links/models"
	"net/http"
	"time"
)

func initializeEmailTokenRoutes() {
	http.HandleFunc("/____reserved/api/verify_email", signupEmailVerificationHandler)
}

func signupEmailVerificationHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	userEmail := r.URL.Query().Get("email")
	if userEmail == "" {
		fmt.Println("Email is required")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if token == "" {
		fmt.Println("Token is required")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fullUser, err := models.GetUserByVerificationTokenAndEmail(token, userEmail)
	if err != nil {
		fmt.Println("No Valid user found with token and email", err)
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
	if fullUser.VerificationToken != token {
		fmt.Println("Verification token mismatch: ", fullUser.VerificationToken, token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if fullUser.VerificationTokenExpiresAt.Before(time.Now()) {
		fmt.Println("Verification token expired")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fullUser.VerifiedEmail = true
	models.SetEmailToVerified(fullUser.Id)
	refreshCookie(fullUser, w)
	// redirect to home page
	http.Redirect(w, r, "/____reserved/verified_email", http.StatusFound)
}
