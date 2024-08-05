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
	user, err := auth.AttemptCookieDecryption(rawCookieHeader)
	if err != nil {
		return common.TrimmedUser{}, err
	}
	return user, nil
}
func getVerifiedUserInCookies(r *http.Request) (common.TrimmedUser, error) {
	user, err := getUserInCookies(r)
	if err != nil {
		return common.TrimmedUser{}, err
	}
	if !user.VerifiedEmail {
		return common.TrimmedUser{}, errors.New("user not verified")
	}
	return user, nil
}
func attemptLogin(w http.ResponseWriter, userId int64, plainTextPassword string) {
	attemptedCookie, err := auth.AttemptLoginAndGetCookie(userId, plainTextPassword)
	if err != nil {
		http.Error(w, "invalid login, no user with that email or password", http.StatusBadRequest)
		return
	}
	// Write the cookie
	w.Header().Set("Set-Cookie", attemptedCookie)
	w.Write([]byte("OK"))
}
func refreshCookie(user common.User, w http.ResponseWriter) {
	cookie, err := auth.CreateNewCookie(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Set-Cookie", cookie)
}
