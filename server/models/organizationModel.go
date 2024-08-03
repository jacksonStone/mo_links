package models

import (
	"errors"
	"mo_links/common"
	"mo_links/db"
)

func GetMatchingOrganizations(userId int64) ([]common.Organization, error) {
	if userId == 0 {
		return []common.Organization{}, errors.New("userId must be defined")
	}
	return db.DbGetMatchingOrganizations(userId)
}

func CreateOrganization(name string, userId int64) error {
	err := ValidOrganizationName(name)
	if err != nil {
		return err
	}
	org, _ := db.DbGetOrganizationByNameAndCreator(name, userId)
	if org.Id != 0 {
		return errors.New("you have already created an organization with this name")
	}

	// Create the organization
	err = db.DbCreateOrganizationAndOwnerMembership(name, userId)
	if err != nil {
		return err
	}
	return nil
}

func AssignMemberToOrganization(userId int64, role string, organizationId int64) error {
	return db.DbAssignMemberToOrganization(userId, role, organizationId)
}

func GetUsersOrganizationAndRoleForEach(userId int64) ([]common.OrganizationMember, error) {
	return db.DbGetUsersOrganizationAndRoleForEach(userId)
}

func GetOrganizationMembers(organizationId int64) ([]common.OrganizationMember, error) {
	return db.DbGetOrganizationMembers(organizationId)
}

func ValidOrganizationName(name string) error {
	// Name must be 1-255 characters long
	if len(name) == 0 || len(name) > 255 {
		return errors.New("name must be 1-255 characters long")
	}
	return nil
}

func GetUserRoleInOrganization(userId int64, organizationId int64) (string, error) {
	memberships, err := GetUsersOrganizationAndRoleForEach(userId)
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
func GetOrganizationById(organizationId int64) (common.Organization, error) {
	return db.DbGetOrganizationById(organizationId)
}
func RoleCanAddRole(userRole string, targetRole string) bool {
	if targetRole == common.RoleAdmin || targetRole == common.RoleOwner {
		return userRole == common.RoleAdmin || userRole == common.RoleOwner
	}
	if targetRole == common.RoleMember {
		return userRole == common.RoleAdmin || userRole == common.RoleOwner
	}
	return false
}
func RoleCanViewMembers(userRole string) bool {
	return userRole == common.RoleAdmin || userRole == common.RoleOwner
}
