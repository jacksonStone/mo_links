package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var stmtGetMatchingLinks *sql.Stmt
var stmtAddLink *sql.Stmt

func initializeDB() {
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
	stmtAddLink = prepareAddLinkStmt()
	stmtGetMatchingLinks = prepareGetMatchingLinksStmt()
}

func prepareAddLinkStmt() *sql.Stmt {
	stmtAddLink, err := db.Prepare(`
	INSERT INTO mo_links_entries (url, name, created_by_user_id) VALUES (?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	return stmtAddLink
}

func prepareGetMatchingLinksStmt() *sql.Stmt {
	stmtGetMatchingLinks, err := db.Prepare(`
	SELECT url, organization_id FROM mo_links_entries 
	WHERE (
		created_by_user_id = ? 
		OR 
		organization_id IN (
			SELECT organization_id FROM mo_links_organization_memberships WHERE user_id = ?
		)
	) AND name = ? ORDER BY created_at DESC`)
	if err != nil {
		log.Fatal(err)
	}
	return stmtGetMatchingLinks
}

func dbGetMatchingLinks(userId int32, name string) ([]string, error) {
	rows, err := stmtGetMatchingLinks.Query(userId, userId, name)
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

func dbAddLink(url string, name string, userId int32) error {
	_, err := stmtAddLink.Exec(url, name, userId)
	if err != nil {
		return err
	}
	return nil
}
