package models

import (
	"errors"
	"mo_links/auth"
	"mo_links/common"
	"mo_links/db"
	"time"
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

func SignupUser(email string, password string) error {
	userId, _ := db.DbGetUserByEmail(email)
	if userId != 0 {
		return errors.New("user with that email alreadyå exists")
	}
	salt := auth.GenerateSalt()
	verificationToken := auth.GenerateSalt() // Get a different random string for the verification token so that the "actual" salt will not be sent over email
	hashedPassword := auth.GetHashedPasswordFromRawTextPassword(password, salt)
	verificationExpiration := int64(time.Now().Add(7 * 24 * time.Hour).Unix())
	// Create the user
	return db.DbSignupUser(email, hashedPassword, salt, verificationToken, verificationExpiration)
}

func GetUserIdByEmail(email string) (int64, error) {
	userId, err := db.DbGetUserByEmail(email)
	if err != nil {
		return 0, err
	}
	return userId, nil
}
func GetUserById(id int64) (common.User, error) {
	user, err := db.DbGetUser(id)
	if err != nil {
		return common.User{}, err
	}
	return user, nil
}
