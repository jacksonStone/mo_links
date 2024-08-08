package http_server

import (
	"encoding/json"
	"mo_links/db"
	"mo_links/models"
	"net/http"
)

type CreateOrganizationRequest struct {
	Name string `json:"name"`
}
type AssignMemberRequest struct {
	UserId         int64  `json:"userId"`
	OrganizationId int64  `json:"organizationId"`
	Role           string `json:"role"`
}
type GetOrganizationMembersRequest struct {
	OrganizationId int64 `json:"organizationId"`
}
type UpdateActiveOrganizationRequest struct {
	OrganizationId int64 `json:"organizationId"`
}

func initializeOrganizationRoutes() {
	http.HandleFunc("/____reserved/api/organizations", organizationsEndpoint)
	http.HandleFunc("/____reserved/api/organization/make_active", makeActiveOrganizationEndpoint)
	http.HandleFunc("/____reserved/api/organization/create", createOrganizationEndpoint)
	http.HandleFunc("/____reserved/api/organization/members", getOrganizationMembersEndpoint)
}

func organizationsEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	organizations, err := models.GetMatchingOrganizations(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organizations)
}

func createOrganizationEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request CreateOrganizationRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = models.CreateOrganization(request.Name, user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Refresh the cookie to update the active org within the cookie
	fullUser, err := models.GetUserById(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	refreshCookie(fullUser, w)

	w.WriteHeader(http.StatusCreated)
}

func makeActiveOrganizationEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var request UpdateActiveOrganizationRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if request.OrganizationId == 0 {
		http.Error(w, "must provide target organization ID", http.StatusBadRequest)
		return
	}
	organizations, err := models.GetMatchingOrganizations(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If we are in that org, we can assign it as our active org
	for _, org := range organizations {
		if org.Id == request.OrganizationId {
			err = db.DbSetUserActiveOrganization(user.Id, request.OrganizationId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Refresh login cookie
			fullUser, err := models.GetUserById(user.Id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			refreshCookie(fullUser, w)
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	http.Error(w, "user is not a member of that organization", http.StatusForbidden)
}

func getOrganizationMembersEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request GetOrganizationMembersRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	role, err := models.GetUserRoleInOrganization(user.Id, request.OrganizationId)

	if err != nil {
		http.Error(w, "Error checking user permissions", http.StatusInternalServerError)
		return
	}
	if role == "" {
		http.Error(w, "User is not a member of this organization", http.StatusForbidden)
		return
	}

	if !models.RoleCanViewMembers(role) {
		http.Error(w, "User is not authorized to view members of this organization", http.StatusForbidden)
		return
	}

	members, err := models.GetOrganizationMembers(request.OrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
