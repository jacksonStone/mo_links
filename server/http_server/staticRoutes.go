package http_server

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed static
var static embed.FS

func initializeStaticRoutes() {

	http.HandleFunc("/____reserved/create_organization", serveCreateOrganizationPage)
	http.HandleFunc("/____reserved/edit_organization", serveEditOrganizationPage)
	http.HandleFunc("/____reserved/privacy_policy", getPrivacyPolicyEndpoint)
	http.HandleFunc("/____reserved/login_page", loginPageEndpoint)
	http.HandleFunc("/____reserved/static/", serveStaticFiles)
	http.HandleFunc("/____reserved/verified_email", serveVerifiedEmailPage)
	http.HandleFunc("/____reserved/get_started", serveGetStartedPage)
	http.HandleFunc("/favicon.ico", faviconEndpoint)

}
func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	}
	if strings.HasSuffix(r.URL.Path, ".js") {
		w.Header().Set("Content-Type", "text/javascript")
	}
	if strings.HasSuffix(r.URL.Path, ".png") {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Header().Set("Content-Type", "image/png")
	}
	if strings.HasSuffix(r.URL.Path, ".svg") {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Header().Set("Content-Type", "image/svg+xml")
	}
	returnStaticFile(w, strings.TrimPrefix(r.URL.Path, "/____reserved/"))
}
func returnStaticFile(w http.ResponseWriter, path string) {
	bytes, err := static.ReadFile(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func serveVerifiedEmailPage(w http.ResponseWriter, r *http.Request) {
	returnStaticFile(w, "static/verified_email.html")
}

func faviconEndpoint(w http.ResponseWriter, r *http.Request) {
	// Set appropriate content type
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	returnStaticFile(w, "static/logo-16.png")
}
func ServeHomePage(w http.ResponseWriter) {
	returnStaticFile(w, "static/index.html")
}

func serveGetStartedPage(w http.ResponseWriter, r *http.Request) {
	returnStaticFile(w, "static/get_started.html")
}

func serveCreateOrganizationPage(w http.ResponseWriter, r *http.Request) {
	returnStaticFile(w, "static/create_organization.html")
}

func serveEditOrganizationPage(w http.ResponseWriter, r *http.Request) {
	returnStaticFile(w, "static/edit_organization.html")
}

func getPrivacyPolicyEndpoint(w http.ResponseWriter, r *http.Request) {
	returnStaticFile(w, "static/privacy_policy.html")
}

func loginPageEndpoint(w http.ResponseWriter, r *http.Request) {
	returnStaticFile(w, "static/login.html")
}
