package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed static
var static embed.FS
var db *sql.DB

type AddLinkRequest struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

func main() {
	rootPath := "./"
	dbPath := "mo_links.db"
	_, err := os.Stat(rootPath + dbPath)
	if os.IsNotExist(err) {
		rootPath = "../../sqlite_wrapper/migrator/"
	}
	path := rootPath + dbPath
	fmt.Println("path: " + path)

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/____reserved/privacy_policy", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := static.ReadFile("static/privacy_policy.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
	})

	http.HandleFunc("/____reserved/api/add", func(w http.ResponseWriter, r *http.Request) {
		var request AddLinkRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		userId := 1 // TODO localize to the user
		_, err = db.Exec(`INSERT INTO mo_links_entries (url, name, created_by_user_id) VALUES (?, ?, ?)`, request.Url, request.Name, userId)
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
	userId := 1
	rows, err := db.Query(`
	SELECT url, organization_id FROM mo_links_entries 
	WHERE (
		created_by_user_id = ? 
		OR 
		organization_id IN (
			SELECT organization_id FROM mo_links_organization_memberships WHERE user_id = ?
		)
	) AND name = ?`, userId, userId, path)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()
	var links []string
	for rows.Next() {
		var url string
		var organizationId int32
		rows.Scan(&url, &organizationId)
		links = append(links, url)
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
