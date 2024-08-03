package http_server

import (
	"encoding/json"
	"mo_links/models"
	"net/http"
	"regexp"
)

type SendInviteRequest struct {
	OrganizationId int64  `json:"organizationId"`
	Email          string `json:"email"`
	Message        string `json:"message"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func initializeInvitesRoute() {
	http.HandleFunc("/____reserved/api/send_invite", sendInviteEndpoint)
	http.HandleFunc("/____reserved/api/accept_invite", acceptInviteEndpoint)
}

func validEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func sendInviteEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req SendInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	organizationId := req.OrganizationId
	if organizationId == 0 {
		http.Error(w, "invalid Organization Id", http.StatusBadRequest)
		return
	}
	if req.Email == "" || !validEmail(req.Email) {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	organization, err := models.GetOrganizationById(organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	role, err := models.GetUserRoleInOrganization(user.Id, organizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if organization.IsPersonal {
		http.Error(w, "you cannot send an invite for a personal organization", http.StatusBadRequest)
		return
	}

	if role != "admin" && role != "owner" {
		http.Error(w, "unauthorized to send invites for organization", http.StatusUnauthorized)
		return
	}
	// TODO Send Email

}
func acceptInviteEndpoint(w http.ResponseWriter, r *http.Request) {

}
