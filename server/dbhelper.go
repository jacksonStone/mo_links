package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitializeDB() {
	rootPath := "./"
	dbPath := "mo_links.db"
	_, err := os.Stat(rootPath + dbPath)
	if os.IsNotExist(err) {
		rootPath = "../../sqlite_wrapper/migrator/"
	}
	path := rootPath + dbPath

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DB initialized: " + path)
}

func GetMatchingLinks(userId int32, path string) ([]string, error) {
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

func AddLink(url string, name string, userId int32) error {
	_, err := db.Exec(`
		INSERT INTO mo_links_entries (url, name, created_by_user_id) VALUES (?, ?, ?)`, url, name, userId)
	if err != nil {
		return err
	}
	return nil
}
