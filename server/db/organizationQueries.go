package db

import (
	"database/sql"
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

func matchingOrganizationsStmt() *sql.Stmt {
	return getQuery(`
    SELECT o.id, o.name, o.is_personal, o.created_by_user_id FROM mo_links_organizations o
    JOIN mo_links_organization_memberships m ON o.id = m.organization_id
    WHERE m.user_id = ?`)
}

func organizationByNameAndCreatorStmt() *sql.Stmt {
	return getQuery(`
    SELECT id, name, is_personal, created_by_user_id FROM mo_links_organizations WHERE name = ? AND created_by_user_id = ?`)
}

func createOrganizationStmt() *sql.Stmt {
	return getQuery(`
    INSERT INTO mo_links_organizations (name, created_by_user_id, is_personal) VALUES (?, ?, ?)`)
}

func assignMemberToOrganizationStmt() *sql.Stmt {
	return getQuery(`
    INSERT INTO mo_links_organization_memberships (user_id, role, organization_id) VALUES (?, ?, ?)`)
}

func getOrganizationMembersStmt() *sql.Stmt {
	return getQuery(`
    SELECT m.organization_id, o.name, u.id, u.email, m.role 
    FROM mo_links_organization_memberships m
    JOIN mo_links_users u ON m.user_id = u.id
    JOIN mo_links_organizations o ON m.organization_id = o.id
    WHERE m.organization_id = ?`)
}

func getUsersOrganizationAndRoleForEachStmt() *sql.Stmt {
	return getQuery(`
    SELECT o.id, o.name, m.user_id, u.email, m.role
    FROM mo_links_organizations o
    JOIN mo_links_organization_memberships m ON o.id = m.organization_id
    JOIN mo_links_users u ON m.user_id = u.id
    WHERE m.user_id = ?`)
}
