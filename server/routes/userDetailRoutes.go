package routes

import (
	"encoding/json"
	"mo_links/auth"
	"mo_links/models"
	"net/http"
)

func InitUserDetailRoutes() {
	http.HandleFunc("/____reserved/api/me", meEndpoint)
}

func meEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userDetails, err := models.GetUserDetails(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userDetails)
}
