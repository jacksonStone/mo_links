package http_server

import (
	"encoding/json"
	"errors"
	"mo_links/auth"
	"mo_links/common"
	"net/http"
	"strconv"
)

func getUserInCookies(r *http.Request) (common.TrimmedUser, error) {
	// Log all cookies
	var authCookie string
	for _, cookie := range r.Cookies() {
		if cookie.Name == auth.Auth.CookieName {
			authCookie = cookie.Value
		}
	}
	if authCookie == "" {
		return common.TrimmedUser{}, errors.New("no auth cookie found")
	}
	// Get raw cookie header
	rawCookieHeader := r.Header.Get("Cookie")
	if rawCookieHeader == "" {
		return common.TrimmedUser{}, errors.New("no raw cookie header found")
	}

	decryptedCookie, err := auth.Auth.AttemptCookieDecryption(rawCookieHeader)
	if err != nil {
		return common.TrimmedUser{}, err
	}
	// parse cookie as json
	var user common.TrimmedUser
	err = json.Unmarshal([]byte(decryptedCookie), &user)
	if err != nil {
		return common.TrimmedUser{}, err
	}
	return user, nil
}
func attemptLogin(w http.ResponseWriter, userId int64, plainTextPassword string) {
	attemptedCookie, err := auth.Auth.AttemptLoginAndGetCookie(strconv.Itoa(int(userId)), plainTextPassword)
	if err != nil {
		http.Error(w, "invalid login, no user with that email or password", http.StatusBadRequest)
		return
	}
	// Write the cookie
	w.Header().Set("Set-Cookie", attemptedCookie)
	w.Write([]byte("OK"))
}
func refreshCookie(user common.User, w http.ResponseWriter) {
	cookie, err := auth.Auth.CreateNewCookie(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Set-Cookie", cookie)
}
