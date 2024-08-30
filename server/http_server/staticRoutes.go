package http_server

import (
	"embed"
	"fmt"
	"net/http"
	"os"
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
	http.HandleFunc("/____reserved/hop", serveHopPage)

	http.HandleFunc("/favicon.ico", faviconEndpoint)

}
func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
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
func pathIsTemplatable(path string) bool {
	return strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".js")
}
func getFileContents(path string) ([]byte, error) {
	if os.Getenv("NODE_ENV") != "development" {
		return static.ReadFile(path)
	} else {
		return os.ReadFile("http_server/" + path)
	}
}

func getParsedFileContents(path string) ([]byte, error) {
	bytes, err := getFileContents(path)
	if err != nil {
		return nil, err
	}
	strContent := string(bytes)

	// Define your placeholders and their replacements here
	replacements := map[string]string{
		"{{STRIPE_PUBLIC_KEY}}": os.Getenv("MOLINKS_STRIPE_PUBLIC_KEY"),
	}

	// Replace placeholders
	for placeholder, replacement := range replacements {
		strContent = strings.ReplaceAll(strContent, placeholder, replacement)
	}

	// Convert back to []byte
	return []byte(strContent), nil
}

var cache = make(map[string][]byte)

func getCacheableFileContents(path string) ([]byte, error) {
	if pathIsTemplatable(path) {
		return getCacheableParsedFileContents(path)
	}
	if os.Getenv("NODE_ENV") == "development" || strings.Contains(path, ".png") || strings.Contains(path, ".svg") {
		return getFileContents(path)
	}
	if cached, ok := cache[path]; ok {
		return cached, nil
	}
	bytes, err := getFileContents(path)
	if err != nil {
		return nil, err
	}
	cache[path] = bytes
	return bytes, nil
}

var parsedCache = make(map[string][]byte)

func getCacheableParsedFileContents(path string) ([]byte, error) {
	if os.Getenv("NODE_ENV") == "development" {
		return getParsedFileContents(path)
	}
	if cached, ok := parsedCache[path]; ok {
		return cached, nil
	}
	bytes, err := getParsedFileContents(path)
	if err != nil {
		return nil, err
	}
	parsedCache[path] = bytes
	return bytes, nil
}

func returnStaticFile(w http.ResponseWriter, path string) {
	bytes, err := getCacheableFileContents(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func serveVerifiedEmailPage(w http.ResponseWriter, r *http.Request) {
	returnStaticFile(w, "static/verified_email.html")
}

func serveHopPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("serving hop page")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	returnStaticFile(w, "static/hop.html")
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
