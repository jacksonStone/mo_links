package http_server

import (
	"encoding/json"
	"mo_links/models"
	"net/http"
	"regexp"
)

type CreateInviteRequest struct {
	OrganizationId int64  `json:"organizationId"`
	InviteeEmail   string `json:"inviteeEmail"`
	EmailMessage   string `json:"emailMessage"`
}
type AcceptInviteRequest struct {
	Token string `json:"token"`
}
type GetOrganizationInvitesRequest struct {
	OrganizationId int64 `json:"organizationId"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func initializeInvitesRoute() {
	http.HandleFunc("/____reserved/api/send_invite", sendInviteEndpoint)
	http.HandleFunc("/____reserved/api/accept_invite", acceptInviteEndpoint)
	http.HandleFunc("/____reserved/api/get_organization_invites", getOrganizationInvitesEndpoint)
}

func validEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func sendInviteEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var req CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	organizationId := req.OrganizationId
	if organizationId == 0 {
		http.Error(w, "invalid Organization Id", http.StatusBadRequest)
		return
	}
	if req.InviteeEmail == "" || !validEmail(req.InviteeEmail) {
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
	err = models.CreateInvite(organizationId, req.InviteeEmail, req.EmailMessage, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)

}
func acceptInviteEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getUserInCookies(r)
	if err != nil {
		redirectToLoginWithNext(w, r)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token in query parameters", http.StatusBadRequest)
		return
	}

	// DO SOMETHING HERE WITH URL ENCODINGS
	err = models.AcceptInvite(token, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !user.VerifiedEmail {
		user.VerifiedEmail = true
		models.SetEmailToVerified(user.Id)
	}
	fullUser, err := models.GetUserById(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	refreshCookie(fullUser, w)

	w.WriteHeader(http.StatusOK)
}

func getOrganizationInvitesEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var request GetOrganizationInvitesRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	role, err := models.GetUserRoleInOrganization(user.Id, request.OrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if role != "admin" && role != "owner" {
		http.Error(w, "unauthorized to view invites for organization", http.StatusUnauthorized)
		return
	}
	invites, err := models.GetOrganizationInvites(request.OrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invites)
}
