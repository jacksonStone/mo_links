package main

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed static
var static embed.FS

type AddLinkRequest struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

func main() {
	InitializeDB()
	http.HandleFunc("/____reserved/privacy_policy", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := static.ReadFile("static/privacy_policy.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	})

	http.HandleFunc("/____reserved/api/add", func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		// Handle preflight request
		if r.Method == http.MethodOptions {
			return
		}
		var request AddLinkRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		userId := int32(1)
		err = AddLink(request.Url, request.Name, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		links, err := decodeLink(r)
		if err != nil {
			// TODO improve error handling
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(links) == 0 {
			// TODO Open up the home page to create a new link
			serveHomePage(w)
			return
		}
		if len(links) > 1 {
			// TODO Support multiple definitions
			http.NotFound(w, r)
			return
		}
		link := links[0]
		// Redirect user to link, but don't cache the result
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.Redirect(w, r, link, http.StatusFound)
	})

	http.ListenAndServe(":3003", nil)
}

func decodeLink(r *http.Request) ([]string, error) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		return []string{}, nil
	}
	userId := int32(1)
	links, err := GetMatchingLinks(userId, path)
	if err != nil {
		return []string{}, err
	}
	return links, nil
}

func serveHomePage(w http.ResponseWriter) {
	bytes, err := static.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
