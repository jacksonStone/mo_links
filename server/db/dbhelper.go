package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var queryMapping map[string]*sql.Stmt = map[string]*sql.Stmt{}

func InitializeDB() {
	rootPath := "./"
	dbPath := "mo_links.db"
	_, err := os.Stat(rootPath + dbPath)
	if os.IsNotExist(err) {
		// Local development
		rootPath = "../../sqlite_wrapper/migrator/"
	}
	path := rootPath + dbPath

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("DB initialized: " + path)
	initializeUserQueries()
	initializeOrganizationQueries()
	initializeLinkQueries()
	initializeInviteQueries()
	fmt.Println("All DB stmts prepared")

}

func getQuery(query string) *sql.Stmt {
	stmt, ok := queryMapping[query]
	if !ok {
		stmt, err := db.Prepare(query)
		if err != nil {
			log.Fatal("error creating query:\n" + query + "\nError: " + err.Error())
		}
		queryMapping[query] = stmt
		return stmt
	}
	return stmt
}
