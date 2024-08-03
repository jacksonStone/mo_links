package models

import (
	"mo_links/common"
	"mo_links/db"
)

func GetUserDetails(trimmedUser common.TrimmedUser) (common.UserDetails, error) {
	var user common.UserDetails
	user.Id = trimmedUser.Id
	user.Email = trimmedUser.Email
	user.ActiveOrganizationId = trimmedUser.ActiveOrganizationId
	memberships, err := GetUsersOrganizationAndRoleForEach(trimmedUser.Id)
	if err != nil {
		return common.UserDetails{}, err
	}
	user.Memberships = memberships
	user.MoLinks, err = db.DbGetUserMoLinks(trimmedUser.Id, trimmedUser.ActiveOrganizationId)
	if err != nil {
		return common.UserDetails{}, err
	}
	return user, nil
}
