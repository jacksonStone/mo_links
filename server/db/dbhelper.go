package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"mo_links/common"

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
func DbGetUserMoLinks(userId int64, organizationId int64) ([]common.MoLink, error) {
	rows, err := getUserMoLinksStmt.Query(userId, organizationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var moLinks []common.MoLink
	for rows.Next() {
		var moLink common.MoLink
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
func DbSetUserActiveOrganization(userId int64, organizationId int64) error {
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
func DbGetUser(userId int64) (common.User, error) {
	row := getUserStmt.QueryRow(userId)
	var user common.User
	err := row.Scan(&user.Id, &user.Email, &user.HashedPassword, &user.Salt, &user.ActiveOrganizationId)
	if err != nil {
		return common.User{}, err
	}
	return user, nil
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

func DbIncrementViewCountOfLink(organizationId int64, name string) error {
	_, err := stmtIncrementViewCountOfLink.Exec(organizationId, name)
	return err
}

func DbGetOrganizationByNameAndCreator(name string, userId int64) (common.Organization, error) {
	row := stmtGetOrganizationByNameAndCreator.QueryRow(name, userId)
	var organization common.Organization
	err := row.Scan(&organization.Id, &organization.Name, &organization.IsPersonal, &organization.CreatedByUserId)
	if err != nil {
		return common.Organization{}, err
	}
	return organization, nil
}

func DbGetMatchingOrganizations(userId int64) ([]common.Organization, error) {
	rows, err := stmtGetMatchingOrganizations.Query(userId)
	if err != nil {
		return []common.Organization{}, err
	}
	defer rows.Close()
	var organizations []common.Organization
	for rows.Next() {
		var org common.Organization
		err := rows.Scan(&org.Id, &org.Name, &org.IsPersonal, &org.CreatedByUserId)
		if err != nil {
			return []common.Organization{}, err
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
	_, err = tx.Stmt(stmtAssignMemberToOrganization).Exec(userId, common.RoleOwner, orgId)
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

func DbCreateOrganizationAndOwnerMembership(name string, userId int64) error {
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

func DbAssignMemberToOrganization(userId int64, role string, organizationId int64) error {
	_, err := stmtAssignMemberToOrganization.Exec(userId, role, organizationId)
	if err != nil {
		return err
	}
	return nil
}

func DbGetOrganizationMembers(organizationId int64) ([]common.OrganizationMember, error) {
	rows, err := stmtGetOrganizationMembers.Query(organizationId)
	if err != nil {
		return []common.OrganizationMember{}, err
	}
	defer rows.Close()
	var members []common.OrganizationMember
	for rows.Next() {
		var member common.OrganizationMember
		err := rows.Scan(&member.OrganizationId, &member.OrganizationName, &member.UserId, &member.UserEmail, &member.UserRole)
		if err != nil {
			return []common.OrganizationMember{}, err
		}
		members = append(members, member)
	}
	return members, nil
}
func DbGetUsersOrganizationAndRoleForEach(userId int64) ([]common.OrganizationMember, error) {
	rows, err := stmtGetUsersOrganizationAndRoleForEach.Query(userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []common.OrganizationMember
	for rows.Next() {
		var member common.OrganizationMember
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

func DbSignupUser(email, passwordHash, passwordSalt, verificationToken string, verificationExperation int64) error {
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
	txCreateOrganizationAndOwnerMembership(tx, common.OrgNamePersonal, userId, true)

	return tx.Commit()
}

// TODO You ar emigrate user methods to the user file

func DbGetUserByEmail(email string) (int64, error) {
	row := stmtGetUserByEmail.QueryRow(email)
	var userId int64
	err := row.Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func DbGetOrganizationById(organizationId int64) (common.Organization, error) {
	row := stmtGetOrganizationById.QueryRow(organizationId)
	var organization common.Organization
	err := row.Scan(&organization.Id, &organization.Name, &organization.IsPersonal, &organization.CreatedByUserId)
	if err != nil {
		return common.Organization{}, err
	}
	return organization, nil
}

func DbGetMatchingLinks(organizationId int64, name string) ([]string, error) {
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

func DbAddLink(url string, name string, userId int64, activeOrganizationId int64) error {
	_, err := stmtAddLink.Exec(url, name, userId, activeOrganizationId)
	if err != nil {
		return err
	}
	return nil
}
