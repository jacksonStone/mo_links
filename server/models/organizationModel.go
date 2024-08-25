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
func ChangeUserRoleInOrganization(userId int64, organizationId int64, newRole string) error {
	return db.DbChangeUserRoleInOrganization(userId, organizationId, newRole)
}
func RemoveUserFromOrganization(userId int64, organizationId int64) error {
	return db.DbRemoveUserFromOrganization(userId, organizationId)
}
func RoleCanChangeMemberRole(userRole string, currentUserTargetRole string, targetRole string) bool {
	if targetRole != common.RoleMember && targetRole != common.RoleAdmin && targetRole != common.RoleOwner {
		// Not a valid role to change to
		return false
	}
	// For now only owners can change user roles
	return userRole == common.RoleOwner
}
func RoleCanRemoveUserOfRole(userRole string, targetRole string) bool {
	if targetRole == common.RoleOwner {
		//Cannot remove owners.
		return false
	}
	if userRole == common.RoleOwner {
		// Owners can remove anyone otherwise
		return true
	}
	// Admins can remove members
	return userRole == common.RoleAdmin && targetRole == common.RoleMember
}
func RoleCanRemoveLink(userRole string) bool {
	return userRole == common.RoleOwner || userRole == common.RoleAdmin
}
func RoleCanUpdateLink(userRole string) bool {
	return userRole == common.RoleOwner || userRole == common.RoleAdmin
}
func RoleCanViewMembers(userRole string) bool {
	return userRole == common.RoleAdmin || userRole == common.RoleOwner
}
