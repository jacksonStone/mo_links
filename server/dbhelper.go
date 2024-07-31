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
var stmtGetUser *sql.Stmt
var stmtGetUserByEmail *sql.Stmt
var stmtSignupUserByEmail *sql.Stmt
var stmtGetMatchingOrganizations *sql.Stmt
var stmtCreateOrganization *sql.Stmt
var stmtAssignMemberToOrganization *sql.Stmt
var stmtGetOrganizationMembers *sql.Stmt
var stmtGetOrganizationByNameAndCreator *sql.Stmt
var stmtGetUsersOrganizationAndRoleForEach *sql.Stmt

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
	stmtGetUser = prepareGetUserStmt()
	stmtGetUserByEmail = prepareGetUserByEmailStmt()
	stmtSignupUserByEmail = prepareSignupUserByEmail()
	stmtGetMatchingOrganizations = prepareGetMatchingOrganizationsStmt()
	stmtCreateOrganization = prepareCreateOrganizationStmt()
	stmtAssignMemberToOrganization = prepareAssignMemberToOrganizationStmt()
	stmtGetOrganizationMembers = prepareGetOrganizationMembersStmt()
	stmtGetOrganizationByNameAndCreator = prepareGetOrganizationByNameAndCreatorStmt()
	stmtGetUsersOrganizationAndRoleForEach = prepareGetUsersOrganizationAndRoleForEachStmt()
	fmt.Println("All DB stmts prepared")

}

func prepareAddLinkStmt() *sql.Stmt {
	stmtAddLink, err := db.Prepare(`
	INSERT INTO mo_links_entries (url, name, created_by_user_id) VALUES (?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	return stmtAddLink
}

func prepareSignupUserByEmail() *sql.Stmt {
	stmtAddLink, err := db.Prepare(`
	INSERT INTO mo_links_users (email, password_hash, password_salt, verification_token, verification_token_expires_at) VALUES (?, ?, ?, ?, ?)`)
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

func prepareGetUserStmt() *sql.Stmt {
	stmtGetUser, err := db.Prepare(`
	SELECT id, email, password_hash, password_salt FROM mo_links_users WHERE id = ? LIMIT 1`)
	if err != nil {
		log.Fatal(err)
	}
	return stmtGetUser
}

func prepareGetUserByEmailStmt() *sql.Stmt {
	stmtGetUserByEmail, err := db.Prepare(`
	SELECT id FROM mo_links_users WHERE email = ? LIMIT 1`)
	if err != nil {
		log.Fatal(err)
	}
	return stmtGetUserByEmail
}

func prepareGetMatchingOrganizationsStmt() *sql.Stmt {
	stmt, err := db.Prepare(`
    SELECT o.id, o.name FROM mo_links_organizations o
    JOIN mo_links_organization_memberships m ON o.id = m.organization_id
    WHERE m.user_id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

func prepareGetOrganizationByNameAndCreatorStmt() *sql.Stmt {
	stmt, err := db.Prepare(`
    SELECT id, name FROM mo_links_organizations WHERE name = ? AND created_by_user_id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

func prepareCreateOrganizationStmt() *sql.Stmt {
	stmt, err := db.Prepare(`
    INSERT INTO mo_links_organizations (name, created_by_user_id) VALUES (?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

func prepareAssignMemberToOrganizationStmt() *sql.Stmt {
	stmt, err := db.Prepare(`
    INSERT INTO mo_links_organization_memberships (user_id, role, organization_id) VALUES (?, ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

func prepareGetOrganizationMembersStmt() *sql.Stmt {
	stmt, err := db.Prepare(`
    SELECT m.organization_id, o.name, u.id, u.email, m.role 
    FROM mo_links_organization_memberships m
    JOIN mo_links_users u ON m.user_id = u.id
    JOIN mo_links_organizations o ON m.organization_id = o.id
    WHERE m.organization_id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}
func prepareGetUsersOrganizationAndRoleForEachStmt() *sql.Stmt {
	stmt, err := db.Prepare(`
    SELECT o.id, o.name, m.user_id, u.email, m.role
    FROM mo_links_organizations o
    JOIN mo_links_organization_memberships m ON o.id = m.organization_id
    JOIN mo_links_users u ON m.user_id = u.id
    WHERE m.user_id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

func dbGetOrganizationByNameAndCreator(name string, userId int32) (Organization, error) {
	row := stmtGetOrganizationByNameAndCreator.QueryRow(name, userId)
	var organization Organization
	err := row.Scan(&organization.Id, &organization.Name)
	if err != nil {
		return Organization{}, err
	}
	return organization, nil
}

func dbGetMatchingOrganizations(userId int32) ([]Organization, error) {
	rows, err := stmtGetMatchingOrganizations.Query(userId)
	if err != nil {
		return []Organization{}, err
	}
	defer rows.Close()
	var organizations []Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(&org.Id, &org.Name)
		if err != nil {
			return []Organization{}, err
		}
		organizations = append(organizations, org)
	}
	return organizations, nil
}

func dbCreateOrganization(name string, userId int32) error {
	_, err := stmtCreateOrganization.Exec(name, userId)
	if err != nil {
		return err
	}
	return nil
}

func dbAssignMemberToOrganization(userId int32, role string, organizationId int32) error {
	_, err := stmtAssignMemberToOrganization.Exec(userId, role, organizationId)
	if err != nil {
		return err
	}
	return nil
}

func dbGetOrganizationMembers(organizationId int32) ([]OrganizationMember, error) {
	rows, err := stmtGetOrganizationMembers.Query(organizationId)
	if err != nil {
		return []OrganizationMember{}, err
	}
	defer rows.Close()
	var members []OrganizationMember
	for rows.Next() {
		var member OrganizationMember
		err := rows.Scan(&member.OrganizationId, &member.OrganizationName, &member.UserId, &member.UserEmail, &member.UserRole)
		if err != nil {
			return []OrganizationMember{}, err
		}
		members = append(members, member)
	}
	return members, nil
}
func dbGetUsersOrganizationAndRoleForEach(userId int32) ([]OrganizationMember, error) {
	rows, err := stmtGetUsersOrganizationAndRoleForEach.Query(userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []OrganizationMember
	for rows.Next() {
		var member OrganizationMember
		err := rows.Scan(&member.OrganizationId, &member.OrganizationName, &member.UserId, &member.UserEmail, &member.UserRole)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func dbGetUser(userId int32) (User, error) {
	row := stmtGetUser.QueryRow(userId)
	var user User
	err := row.Scan(&user.id, &user.email, &user.hashedPassword, &user.salt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func dbSignupUser(email, passwordHash, passwordSalt, verificationToken string, verificationExperation int32) error {
	_, err := stmtSignupUserByEmail.Exec(email, passwordHash, passwordSalt, verificationToken, verificationExperation)
	if err != nil {
		return err
	}
	return nil
}

func dbGetUserByEmail(email string) (int32, error) {
	row := stmtGetUserByEmail.QueryRow(email)
	var userId int32
	err := row.Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
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
