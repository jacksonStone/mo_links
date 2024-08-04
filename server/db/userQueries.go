package db

import (
	"database/sql"
	"fmt"
	"mo_links/common"
)

func initializeUserQueries() {
	getUserStmt()
	setUserActiveOrganizationStmt()
	signupUserByEmailStmt()
	userByEmailStmt()
}

func getUserStmt() *sql.Stmt {
	return getQuery(`
	SELECT id, email, password_hash, password_salt, active_organization_id FROM mo_links_users WHERE id = ? LIMIT 1`)
}
func DbGetUser(userId int64) (common.User, error) {
	row := getUserStmt().QueryRow(userId)
	var user common.User
	err := row.Scan(&user.Id, &user.Email, &user.HashedPassword, &user.Salt, &user.ActiveOrganizationId)
	if err != nil {
		return common.User{}, err
	}
	return user, nil
}

func setUserActiveOrganizationStmt() *sql.Stmt {
	return getQuery(`
		UPDATE mo_links_users SET active_organization_id = ? WHERE id = ?`)
}
func DbSetUserActiveOrganization(userId int64, organizationId int64) error {
	fmt.Println("Setting user:", userId, " to active organization: ", organizationId)
	_, err := setUserActiveOrganizationStmt().Exec(organizationId, userId)
	if err != nil {
		return err
	}
	return nil
}

func signupUserByEmailStmt() *sql.Stmt {
	return getQuery(`
	INSERT INTO mo_links_users (email, password_hash, password_salt, verification_token, verification_token_expires_at) VALUES (?, ?, ?, ?, ?)`)
}
func DbSignupUser(email, passwordHash, passwordSalt, verificationToken string, verificationExperation int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // This will be a no-op if the transaction is committed

	userResult, err := tx.Stmt(signupUserByEmailStmt()).Exec(email, passwordHash, passwordSalt, verificationToken, verificationExperation)
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

func userByEmailStmt() *sql.Stmt {
	return getQuery(`
	SELECT id FROM mo_links_users WHERE email = ? LIMIT 1`)
}
func DbGetUserByEmail(email string) (int64, error) {
	row := userByEmailStmt().QueryRow(email)
	var userId int64
	err := row.Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
}
