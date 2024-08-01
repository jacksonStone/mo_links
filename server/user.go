package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type UserDetails struct {
	Id                   int64
	Email                string
	ActiveOrganizationId int64
	Memberships          []OrganizationMember
	MoLinks              []MoLink
}
type MoLink struct {
	Id             int64
	Name           string
	Url            string
	OrganizationId int64
	CreatedAt      time.Time
	Views          int64
}

var getUserMoLinksStmt *sql.Stmt

func initUserQueries(db *sql.DB) {
	stmt, err := db.Prepare(`
    SELECT id, name, url, organization_id, created_at, views FROM mo_links_entries WHERE created_by_user_id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	getUserMoLinksStmt = stmt
}
func getUserMoLinks(userId int64) ([]MoLink, error) {
	rows, err := getUserMoLinksStmt.Query(userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var moLinks []MoLink
	for rows.Next() {
		var moLink MoLink
		err = rows.Scan(&moLink.Id, &moLink.Name, &moLink.Url, &moLink.OrganizationId, &moLink.CreatedAt, &moLink.Views)
		if err != nil {
			return nil, err
		}
		moLinks = append(moLinks, moLink)
	}
	return moLinks, nil
}
func getUserDetails(trimmedUser TrimmedUser) (UserDetails, error) {
	var user UserDetails
	user.Id = trimmedUser.Id
	user.Email = trimmedUser.Email
	user.ActiveOrganizationId = trimmedUser.ActiveOrganizationId
	memberships, err := getUsersOrganizationAndRoleForEach(trimmedUser.Id)
	if err != nil {
		return UserDetails{}, err
	}
	user.Memberships = memberships
	user.MoLinks, err = getUserMoLinks(trimmedUser.Id)
	if err != nil {
		return UserDetails{}, err
	}
	return user, nil
}
