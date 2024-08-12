package http_server

import (
	"errors"
	"mo_links/auth"
	"mo_links/common"
	"net/http"
)

func getUserInCookies(r *http.Request) (common.TrimmedUser, error) {
	// Get raw cookie header
	rawCookieHeader := r.Header.Get("Cookie")
	if rawCookieHeader == "" {
		return common.TrimmedUser{}, errors.New("no raw cookie header found")
	}
	user, err := getDecryptedToken(rawCookieHeader)
	if err != nil {
		return common.TrimmedUser{}, err
	}
	return user, nil
}
func getDecryptedToken(rawEncryptedToken string) (common.TrimmedUser, error) {
	return auth.AttemptCookieDecryption(rawEncryptedToken)
}
func getVerifiedUserInCookies(r *http.Request) (common.TrimmedUser, error) {
	var user common.TrimmedUser
	if r.URL.Query().Get("token") != "" {
		trimmedUser, err := getDecryptedToken(r.URL.Query().Get("token"))
		if err != nil {
			return common.TrimmedUser{}, err
		}
		user = trimmedUser
	} else {
		trimmedUser, err := getUserInCookies(r)
		if err != nil {
			return common.TrimmedUser{}, err
		}
		user = trimmedUser
	}
	if !user.VerifiedEmail {
		return common.TrimmedUser{}, errors.New("user not verified")
	}
	return user, nil
}
func attemptLogin(r *http.Request, w http.ResponseWriter, userId int64, plainTextPassword string) {
	attemptedCookie, err := auth.AttemptLoginAndGetCookie(userId, plainTextPassword)
	if err != nil {
		http.Error(w, "invalid login, no user with that email or password", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("get_token") == "true" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(attemptedCookie))
		return
	}
	// Write the cookie
	w.Header().Set("Set-Cookie", attemptedCookie)
	w.Write([]byte("OK"))
}
func refreshCookie(user common.User, w http.ResponseWriter) {
	cookie, err := refreshToken(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Set-Cookie", cookie)
}
func refreshToken(user common.User) (string, error) {
	return auth.CreateNewCookie(user)
}
