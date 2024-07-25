package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"unicode"

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

func GetMatchingLinks(userId int32, name string) ([]string, error) {
	rows, err := db.Query(`
	SELECT url, organization_id FROM mo_links_entries 
	WHERE (
		created_by_user_id = ? 
		OR 
		organization_id IN (
			SELECT organization_id FROM mo_links_organization_memberships WHERE user_id = ?
		)
	) AND name = ? ORDER BY created_at DESC`, userId, userId, name)
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
	err := validName(name)
	if err != nil {
		return err
	}
	err = validUrl(url)
	if err != nil {
		return err
	}
	links, err := GetMatchingLinks(userId, name)
	if err != nil {
		return err
	}
	if len(links) > 0 {
		return errors.New("link already exists")
	}
	_, err = db.Exec(`
		INSERT INTO mo_links_entries (url, name, created_by_user_id) VALUES (?, ?, ?)`, url, name, userId)
	if err != nil {
		return err
	}
	return nil
}

func validUrl(url string) error {
	if url == "" {
		return errors.New("url must not be empty")
	}
	// can't be longer than 2000 charecters
	if len(url) > 2000 {
		return errors.New("url must be 2000 characters or less")
	}
	return nil
}

func validName(name string) error {
	// Name must be 1-255 characters long
	if len(name) == 0 || len(name) > 255 {
		return errors.New("name must be 1-255 characters long")
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '-' {
			return errors.New("name must only contain letters, digits, underscores, and hyphens")
		}
	}
	if name == "____reserved" {
		return errors.New("name must not be '____reserved'")
	}
	return nil
}
