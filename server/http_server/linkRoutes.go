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

func initializeLinkRoutes() {
	http.HandleFunc("/____reserved/api/add", addLinkEndpoint)
	http.HandleFunc("/", handleAttemptedMoLink)
}

func handleAttemptedMoLink(w http.ResponseWriter, r *http.Request) {
	// For when running locally without the reverse proxy
	if strings.HasSuffix(r.URL.Path, "/____reserved/_ping") {
		w.Write([]byte("OK"))
		return
	}
	user, err := getUserInCookies(r)
	if err != nil {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.Redirect(w, r, "/____reserved/login_page?next="+url.QueryEscape(r.URL.Path), http.StatusFound)
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
	user, err := getVerifiedUserInCookies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// Allow CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	// Handle preflight request
	if r.Method == http.MethodOptions {
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
