package main

import (
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type Organization struct {
	Id   int32
	Name string
}
type OrganizationMember struct {
	OrganizationId   int32
	UserId           int32
	UserEmail        string
	OrganizationName string
	UserRole         string
}

const (
	RoleAdmin  = "Admin"
	RoleOwner  = "Owner"
	RoleMember = "Member"
)

func getMatchingOrganizations(userId int32) ([]Organization, error) {
	if userId == 0 {
		return []Organization{}, errors.New("userId must be defined")
	}
	return dbGetMatchingOrganizations(userId)
}

func createOrganization(name string, userId int32) error {
	err := validOrganizationName(name)
	if err != nil {
		return err
	}
	org, err := dbGetOrganizationByNameAndCreator(name, userId)
	if err != nil {
		return err
	}
	if org.Id != 0 {
		return errors.New("organization already exists")
	}
	return dbCreateOrganization(name, userId)
}

func assignMemberToOrganization(userId int32, role string, organizationId int32) error {
	return dbAssignMemberToOrganization(userId, role, organizationId)
}

func getUsersOrganizationAndRoleForEach(userId int32) ([]OrganizationMember, error) {
	return dbGetUsersOrganizationAndRoleForEach(userId)
}

func getOrganizationMembers(organizationId int32) ([]OrganizationMember, error) {
	return dbGetOrganizationMembers(organizationId)
}

func validOrganizationName(name string) error {
	// Name must be 1-255 characters long
	if len(name) == 0 || len(name) > 255 {
		return errors.New("name must be 1-255 characters long")
	}
	return nil
}

func getUserRoleInOrganization(userId int32, organizationId int32) (string, error) {
	memberships, err := getUsersOrganizationAndRoleForEach(userId)
	if err != nil {
		return "", err
	}
	for _, org := range memberships {
		if org.OrganizationId == organizationId {
			return org.UserRole, nil
		}
	}
	return "", nil
}
func roleCanAddRole(userRole string, targetRole string) bool {
	if targetRole == RoleAdmin || targetRole == RoleOwner {
		return userRole == RoleAdmin || userRole == RoleOwner
	}
	if targetRole == RoleMember {
		return userRole == RoleAdmin || userRole == RoleOwner
	}
	return false
}
func roleCanViewMembers(userRole string) bool {
	return userRole == RoleAdmin || userRole == RoleOwner
}
