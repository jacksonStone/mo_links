package http_server

import (
	"mo_links/models"
	"net/http"
	"time"
)

func initializeEmailTokenRoutes() {
	http.HandleFunc("/___reserved/api/verify-email", signupEmailVerificationHandler)
}

func signupEmailVerificationHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	user, err := getUserInCookies(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fullUser, err := models.GetUserById(user.Id)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if fullUser.VerifiedEmail {
		// Already Verified - just leave
		w.WriteHeader(http.StatusOK)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if fullUser.VerificationToken != token {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if fullUser.VerificationTokenExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fullUser.VerifiedEmail = true
	models.SetEmailToVerified(user.Id)
	refreshCookie(fullUser, w)
	// redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}
