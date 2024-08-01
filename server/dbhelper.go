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
var stmtGetUserByEmail *sql.Stmt
var stmtSignupUserByEmail *sql.Stmt
var stmtGetMatchingOrganizations *sql.Stmt
var stmtCreateOrganization *sql.Stmt
var stmtAssignMemberToOrganization *sql.Stmt
var stmtGetOrganizationMembers *sql.Stmt
var stmtGetOrganizationByNameAndCreator *sql.Stmt
var stmtGetUsersOrganizationAndRoleForEach *sql.Stmt
var stmtGetOrganizationById *sql.Stmt
var stmtIncrementViewCountOfLink *sql.Stmt

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
	initUserQueries(db)
	stmtAddLink = prepareAddLinkStmt()
	stmtGetMatchingLinks = prepareGetMatchingLinksStmt()
	stmtGetUserByEmail = prepareGetUserByEmailStmt()
	stmtSignupUserByEmail = prepareSignupUserByEmail()
	stmtGetMatchingOrganizations = prepareGetMatchingOrganizationsStmt()
	stmtCreateOrganization = prepareCreateOrganizationStmt()
	stmtAssignMemberToOrganization = prepareAssignMemberToOrganizationStmt()
	stmtGetOrganizationMembers = prepareGetOrganizationMembersStmt()
	stmtGetOrganizationByNameAndCreator = prepareGetOrganizationByNameAndCreatorStmt()
	stmtGetUsersOrganizationAndRoleForEach = prepareGetUsersOrganizationAndRoleForEachStmt()
	stmtGetOrganizationById = prepareGetOrganizationByIdStmt()
	stmtIncrementViewCountOfLink = prepareIncrementViewCountOfLinkStmt()
	fmt.Println("All DB stmts prepared")

}

func prepareGetOrganizationByIdStmt() *sql.Stmt {
	stmtGetOrganizationById, err := db.Prepare(`
	SELECT id, name, is_personal, created_by_user_id FROM mo_links_organizations WHERE id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	return stmtGetOrganizationById
}
func prepareIncrementViewCountOfLinkStmt() *sql.Stmt {
	stmt, err := db.Prepare(`
    UPDATE mo_links_entries SET views = views + 1 
    WHERE organization_id = ? AND name = ?`)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

func prepareAddLinkStmt() *sql.Stmt {
	stmtAddLink, err := db.Prepare(`
	INSERT INTO mo_links_entries (url, name, created_by_user_id, organization_id) VALUES (?, ?, ?, ?)`)
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
	WHERE organization_id = ?
	 AND name = ? ORDER BY created_at DESC`)
	if err != nil {
		log.Fatal(err)
	}
	return stmtGetMatchingLinks
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
    SELECT o.id, o.name, o.is_personal, o.created_by_user_id FROM mo_links_organizations o
    JOIN mo_links_organization_memberships m ON o.id = m.organization_id
    WHERE m.user_id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

func prepareGetOrganizationByNameAndCreatorStmt() *sql.Stmt {
	stmt, err := db.Prepare(`
    SELECT id, name, is_personal, created_by_user_id FROM mo_links_organizations WHERE name = ? AND created_by_user_id = ?`)
	if err != nil {
		log.Fatal(err)
	}
	return stmt
}

func prepareCreateOrganizationStmt() *sql.Stmt {
	stmt, err := db.Prepare(`
    INSERT INTO mo_links_organizations (name, created_by_user_id, is_personal) VALUES (?, ?, ?)`)
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

func dbIncrementViewCountOfLink(organizationId int64, name string) error {
	_, err := stmtIncrementViewCountOfLink.Exec(organizationId, name)
	return err
}

func dbGetOrganizationByNameAndCreator(name string, userId int64) (Organization, error) {
	row := stmtGetOrganizationByNameAndCreator.QueryRow(name, userId)
	var organization Organization
	err := row.Scan(&organization.Id, &organization.Name, &organization.IsPersonal, &organization.CreatedByUserId)
	if err != nil {
		return Organization{}, err
	}
	return organization, nil
}

func dbGetMatchingOrganizations(userId int64) ([]Organization, error) {
	rows, err := stmtGetMatchingOrganizations.Query(userId)
	if err != nil {
		return []Organization{}, err
	}
	defer rows.Close()
	var organizations []Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(&org.Id, &org.Name, &org.IsPersonal, &org.CreatedByUserId)
		if err != nil {
			return []Organization{}, err
		}
		organizations = append(organizations, org)
	}
	return organizations, nil
}
func txCreateOrganizationAndOwnerMembership(tx *sql.Tx, name string, userId int64, isPersonal bool) error {
	// Create the organization
	result, err := tx.Stmt(stmtCreateOrganization).Exec(name, userId, isPersonal)
	if err != nil {
		return err
	}

	orgId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Create the membership with Owner role
	_, err = tx.Stmt(stmtAssignMemberToOrganization).Exec(userId, RoleOwner, orgId)
	if err != nil {
		return err
	}

	// Set the user's active organization to the newly created organization
	_, err = tx.Stmt(setUserActiveOrganizationStmt).Exec(orgId, userId)
	if err != nil {
		return err
	}
	return nil
}

func dbCreateOrganizationAndOwnerMembership(name string, userId int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // This will be a no-op if the transaction is committed

	err = txCreateOrganizationAndOwnerMembership(tx, name, userId, false)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

func dbAssignMemberToOrganization(userId int64, role string, organizationId int64) error {
	_, err := stmtAssignMemberToOrganization.Exec(userId, role, organizationId)
	if err != nil {
		return err
	}
	return nil
}

func dbGetOrganizationMembers(organizationId int64) ([]OrganizationMember, error) {
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
func dbGetUsersOrganizationAndRoleForEach(userId int64) ([]OrganizationMember, error) {
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

func dbSignupUser(email, passwordHash, passwordSalt, verificationToken string, verificationExperation int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // This will be a no-op if the transaction is committed

	userResult, err := tx.Stmt(stmtSignupUserByEmail).Exec(email, passwordHash, passwordSalt, verificationToken, verificationExperation)
	if err != nil {
		return err
	}
	userId, err := userResult.LastInsertId()
	if err != nil {
		return err
	}
	txCreateOrganizationAndOwnerMembership(tx, OrgNamePersonal, userId, true)

	return tx.Commit()
}

// TODO You ar emigrate user methods to the user file

func dbGetUserByEmail(email string) (int64, error) {
	row := stmtGetUserByEmail.QueryRow(email)
	var userId int64
	err := row.Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func dbGetOrganizationById(organizationId int64) (Organization, error) {
	row := stmtGetOrganizationById.QueryRow(organizationId)
	var organization Organization
	err := row.Scan(&organization.Id, &organization.Name, &organization.IsPersonal, &organization.CreatedByUserId)
	if err != nil {
		return Organization{}, err
	}
	return organization, nil
}

func dbGetMatchingLinks(organizationId int64, name string) ([]string, error) {
	fmt.Println("dbGetMatchingLinks", organizationId, name)
	rows, err := stmtGetMatchingLinks.Query(organizationId, name)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()
	var links []string
	for rows.Next() {
		var url string
		var organizationId int64
		rows.Scan(&url, &organizationId)
		links = append(links, url)
	}
	return links, nil
}

func dbAddLink(url string, name string, userId int64, activeOrganizationId int64) error {
	_, err := stmtAddLink.Exec(url, name, userId, activeOrganizationId)
	if err != nil {
		return err
	}
	return nil
}
