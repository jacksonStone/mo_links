package main

import (
	"errors"
)

type Organization struct {
	Id              int64
	Name            string
	IsPersonal      bool
	CreatedByUserId int64
}
type OrganizationMember struct {
	OrganizationId   int64
	UserId           int64
	UserEmail        string
	OrganizationName string
	UserRole         string
}

const (
	RoleAdmin  = "Admin"
	RoleOwner  = "Owner"
	RoleMember = "Member"
)

const (
	OrgNamePersonal = "Personal"
)

func getMatchingOrganizations(userId int64) ([]Organization, error) {
	if userId == 0 {
		return []Organization{}, errors.New("userId must be defined")
	}
	return dbGetMatchingOrganizations(userId)
}

func createOrganization(name string, userId int64) error {
	err := validOrganizationName(name)
	if err != nil {
		return err
	}
	org, _ := dbGetOrganizationByNameAndCreator(name, userId)
	if org.Id != 0 {
		return errors.New("you have already created an organization with this name")
	}

	// Create the organization
	err = dbCreateOrganizationAndOwnerMembership(name, userId)
	if err != nil {
		return err
	}
	return nil
}

func assignMemberToOrganization(userId int64, role string, organizationId int64) error {
	return dbAssignMemberToOrganization(userId, role, organizationId)
}

func getUsersOrganizationAndRoleForEach(userId int64) ([]OrganizationMember, error) {
	return dbGetUsersOrganizationAndRoleForEach(userId)
}

func getOrganizationMembers(organizationId int64) ([]OrganizationMember, error) {
	return dbGetOrganizationMembers(organizationId)
}

func validOrganizationName(name string) error {
	// Name must be 1-255 characters long
	if len(name) == 0 || len(name) > 255 {
		return errors.New("name must be 1-255 characters long")
	}
	return nil
}

func getUserRoleInOrganization(userId int64, organizationId int64) (string, error) {
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
func getOrganizationById(organizationId int64) (Organization, error) {
	return dbGetOrganizationById(organizationId)
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
