package main

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
)

//go:embed static/index.html
var static embed.FS

func main() {
	http.HandleFunc("/____reserved/privacy_policy", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "I will never look at your content. The only thing this chrome extension will do is expand mo/ to https://www.molinks.me/. Everything else is managed within https://www.molinks.me/.")
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
		if link == "" {
			serveHomePage(w)
			return
		}
		fmt.Fprintf(w, "Link: %s", link)
	})

	http.ListenAndServe(":3003", nil)
}
func decodeLink(r *http.Request) ([]string, error) {
	return []string{strings.TrimPrefix(r.URL.Path, "/")}, nil
}
func serveHomePage(w http.ResponseWriter) {
	bytes, err := static.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
