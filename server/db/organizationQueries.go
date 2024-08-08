package db

import (
	"database/sql"
	"mo_links/common"
)

func initializeOrganizationQueries() {
	organizationByIdStmt()
	matchingOrganizationsStmt()
	organizationByNameAndCreatorStmt()
	createOrganizationStmt()
	assignMemberToOrganizationStmt()
	getOrganizationMembersStmt()
	getUsersOrganizationAndRoleForEachStmt()
}

func organizationByIdStmt() *sql.Stmt {
	return getQuery(`
	SELECT id, name, is_personal, created_by_user_id FROM mo_links_organizations WHERE id = ?`)
}
func DbGetOrganizationById(organizationId int64) (common.Organization, error) {
	row := organizationByIdStmt().QueryRow(organizationId)
	var organization common.Organization
	err := row.Scan(&organization.Id, &organization.Name, &organization.IsPersonal, &organization.CreatedByUserId)
	if err != nil {
		return common.Organization{}, err
	}
	return organization, nil
}

func matchingOrganizationsStmt() *sql.Stmt {
	return getQuery(`
    SELECT o.id, o.name, o.is_personal, o.created_by_user_id FROM mo_links_organizations o
    JOIN mo_links_organization_memberships m ON o.id = m.organization_id
    WHERE m.user_id = ?`)
}
func DbGetMatchingOrganizations(userId int64) ([]common.Organization, error) {
	rows, err := matchingOrganizationsStmt().Query(userId)
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

func organizationByNameAndCreatorStmt() *sql.Stmt {
	return getQuery(`
    SELECT id, name, is_personal, created_by_user_id FROM mo_links_organizations WHERE name = ? AND created_by_user_id = ?`)
}
func DbGetOrganizationByNameAndCreator(name string, userId int64) (common.Organization, error) {
	row := organizationByNameAndCreatorStmt().QueryRow(name, userId)
	var organization common.Organization
	err := row.Scan(&organization.Id, &organization.Name, &organization.IsPersonal, &organization.CreatedByUserId)
	if err != nil {
		return common.Organization{}, err
	}
	return organization, nil
}

func assignMemberToOrganizationStmt() *sql.Stmt {
	return getQuery(`
    INSERT INTO mo_links_organization_memberships (user_id, role, organization_id) VALUES (?, ?, ?)`)
}
func DbAssignMemberToOrganization(userId int64, role string, organizationId int64) error {
	_, err := assignMemberToOrganizationStmt().Exec(userId, role, organizationId)
	if err != nil {
		return err
	}
	return nil
}
func createOrganizationStmt() *sql.Stmt {
	return getQuery(`
    INSERT INTO mo_links_organizations (name, created_by_user_id, is_personal) VALUES (?, ?, ?)`)
}
func txCreateOrganizationAndOwnerMembership(tx *sql.Tx, name string, userId int64, isPersonal bool) error {
	// Create the organization
	result, err := tx.Stmt(createOrganizationStmt()).Exec(name, userId, isPersonal)
	if err != nil {
		return err
	}
	orgId, err := result.LastInsertId()
	if err != nil {
		return err
	}
	// Create the membership with Owner role
	_, err = tx.Stmt(assignMemberToOrganizationStmt()).Exec(userId, common.RoleOwner, orgId)
	if err != nil {
		return err
	}
	// Set the user's active organization to the newly created organization
	_, err = tx.Stmt(setUserActiveOrganizationStmt()).Exec(orgId, userId)
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

func getOrganizationMembersStmt() *sql.Stmt {
	return getQuery(`
    SELECT m.organization_id, o.name, u.id, u.email, m.role 
    FROM mo_links_organization_memberships m
    JOIN mo_links_users u ON m.user_id = u.id
    JOIN mo_links_organizations o ON m.organization_id = o.id
    WHERE m.organization_id = ?`)
}
func DbGetOrganizationMembers(organizationId int64) ([]common.OrganizationMember, error) {
	rows, err := getOrganizationMembersStmt().Query(organizationId)
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

func getUsersOrganizationAndRoleForEachStmt() *sql.Stmt {
	return getQuery(`
    SELECT o.id, o.name, m.user_id, u.email, m.role, o.is_personal
    FROM mo_links_organizations o
    JOIN mo_links_organization_memberships m ON o.id = m.organization_id
    JOIN mo_links_users u ON m.user_id = u.id
    WHERE m.user_id = ?`)
}
func DbGetUsersOrganizationAndRoleForEach(userId int64) ([]common.OrganizationMember, error) {
	rows, err := getUsersOrganizationAndRoleForEachStmt().Query(userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []common.OrganizationMember
	for rows.Next() {
		var member common.OrganizationMember
		err := rows.Scan(&member.OrganizationId, &member.OrganizationName, &member.UserId, &member.UserEmail, &member.UserRole, &member.IsPersonal)
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
