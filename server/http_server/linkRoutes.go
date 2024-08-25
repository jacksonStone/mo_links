package http_server

import (
	"encoding/json"
	"mo_links/models"
	"net/http"
	"net/url"
	"strings"
)

type AddLinkRequest struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type RemoveLinkRequest struct {
	Id int64 `json:"id"`
}

type UpdateLinkRequest struct {
	Id  int64  `json:"id"`
	Url string `json:"url"`
}

func initializeLinkRoutes() {
	http.HandleFunc("/____reserved/api/add", addLinkEndpoint)
	http.HandleFunc("/____reserved/api/remove_link", removeLinkEndpoint)
	http.HandleFunc("/____reserved/api/update_link", updateLinkEndpoint)
	http.HandleFunc("/", handleAttemptedMoLink)
}
func redirectToLoginWithNext(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.Redirect(w, r, "/____reserved/login_page?next="+url.QueryEscape(r.URL.Path+"?"+r.URL.RawQuery), http.StatusFound)
}
func handleAttemptedMoLink(w http.ResponseWriter, r *http.Request) {
	// For when running locally without the reverse proxy
	if strings.HasSuffix(r.URL.Path, "/____reserved/_ping") {
		w.Write([]byte("OK"))
		return
	}
	user, err := getUserInCookies(r)
	if err != nil {
		redirectToLoginWithNext(w, r)
		return
	}
	if r.URL.Path == "/" {
		ServeHomePage(w)
		return
	}
	_, err = getVerifiedUserInCookies(r)
	if err != nil {
		// If user has not validated email, redirect to the page telling them this.
		ServeHomePage(w)
		return
	}
	links, err := decodeLink(r, user.ActiveOrganizationId)
	if err != nil {
		// TODO improve error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(links) == 0 {
		// TODO Open up the home page to create a new link
		ServeHomePage(w)
		return
	}
	link := links[0]
	// Redirect user to link, but don't cache the result
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.Redirect(w, r, link, http.StatusFound)
}

func addLinkEndpoint(w http.ResponseWriter, r *http.Request) {
	// Allow CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	// Handle preflight request
	if r.Method == http.MethodOptions {
		return
	}
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var request AddLinkRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = models.AddLink(request.Url, request.Name, user.Id, user.ActiveOrganizationId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusExpectationFailed)
		return
	}
}

func removeLinkEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	var request RemoveLinkRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	link, err := models.GetLink(request.Id)
	if err != nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}
	/** User must be an admin or the creator of the link, and still within the organization */
	role, err := models.GetUserRoleInOrganization(user.Id, link.OrganizationId)
	if err != nil {
		http.Error(w, "User is not within organization", http.StatusUnauthorized)
		return
	}
	if !models.RoleCanRemoveLink(role) && user.Id != link.CreatedByUserId {
		http.Error(w, "User is not authorized to remove this link", http.StatusForbidden)
		return
	}

	err = models.RemoveLink(request.Id)
	if err != nil {
		http.Error(w, "Failed to remove link", http.StatusExpectationFailed)
		return
	}
}

func updateLinkEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	var request UpdateLinkRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	link, err := models.GetLink(request.Id)
	if err != nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}
	/** User must be an admin or the creator of the link */
	role, err := models.GetUserRoleInOrganization(user.Id, link.OrganizationId)
	if err != nil {
		http.Error(w, "User is not within organization", http.StatusUnauthorized)
		return
	}
	if !models.RoleCanUpdateLink(role) && user.Id != link.CreatedByUserId {
		http.Error(w, "User is not authorized to update this link", http.StatusForbidden)
		return
	}
	err = models.UpdateLink(request.Id, request.Url)
	if err != nil {
		http.Error(w, "Failed to update link", http.StatusExpectationFailed)
		return
	}

}

func decodeLink(r *http.Request, organizationId int64) ([]string, error) {

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		return []string{}, nil
	}
	links, err := models.GetMatchingLinks(organizationId, path)
	if err != nil {
		return []string{}, err
	}
	return links, nil
}
