package main

import (
	"database/sql"
	"fmt"
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
var setUserActiveOrganizationStmt *sql.Stmt
var getUserStmt *sql.Stmt

func initUserQueries(db *sql.DB) {
	prepareGetUserMoLinksStmt(db)
	prepareSetUserActiveOrganizationStmt(db)
	prepareGetUserStmt(db)
}
func prepareGetUserMoLinksStmt(db *sql.DB) {
	stmt, err := db.Prepare(`
    SELECT id, name, url, organization_id, created_at, views FROM mo_links_entries WHERE created_by_user_id = ? AND organization_id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	getUserMoLinksStmt = stmt
}
func dbGetUserMoLinks(userId int64, organizationId int64) ([]MoLink, error) {
	rows, err := getUserMoLinksStmt.Query(userId, organizationId)
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
func prepareSetUserActiveOrganizationStmt(db *sql.DB) {
	stmt, err := db.Prepare(`
		UPDATE mo_links_users SET active_organization_id = ? WHERE id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	setUserActiveOrganizationStmt = stmt
}
func dbSetUserActiveOrganization(userId int64, organizationId int64) error {
	fmt.Println("Setting user:", userId, " to active organization: ", organizationId)
	_, err := setUserActiveOrganizationStmt.Exec(organizationId, userId)
	if err != nil {
		return err
	}
	return nil
}
func prepareGetUserStmt(db *sql.DB) {
	stmt, err := db.Prepare(`
	SELECT id, email, password_hash, password_salt, active_organization_id FROM mo_links_users WHERE id = ? LIMIT 1`)
	if err != nil {
		log.Fatal(err)
	}
	getUserStmt = stmt
}
func dbGetUser(userId int64) (User, error) {
	row := getUserStmt.QueryRow(userId)
	var user User
	err := row.Scan(&user.id, &user.email, &user.hashedPassword, &user.salt, &user.activeOrganizationId)
	if err != nil {
		return User{}, err
	}
	return user, nil
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
	user.MoLinks, err = dbGetUserMoLinks(trimmedUser.Id, trimmedUser.ActiveOrganizationId)
	if err != nil {
		return UserDetails{}, err
	}
	return user, nil
}
