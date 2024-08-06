package models

import (
	"errors"
	"mo_links/auth"
	"mo_links/common"
	"mo_links/db"
	"mo_links/utilities"
)

const MaxUnacceptedInvites = 3

func CreateInvite(organizationId int64, inviteeEmail string, emailMessage string, user common.TrimmedUser) error {
	// Check for existing unaccepted invites
	invites, err := db.DbGetOrganizationInvites(organizationId)
	if err != nil {
		return err
	}
	unacceptedCount := 0
	for _, invite := range invites {
		if !invite.Accepted {
			unacceptedCount++
		}
		if invite.InviteeEmail == inviteeEmail && !invite.Accepted {
			return errors.New("an unaccepted invite already exists for this email")
		}
	}

	if unacceptedCount >= MaxUnacceptedInvites {
		return errors.New("organization has reached the maximum number of unaccepted invites")
	}

	token := auth.GenerateUrlSafeToken()

	organization, err := db.DbGetOrganizationById(organizationId)
	if err != nil {
		return err
	}
	utilities.SendInviteEmail(inviteeEmail, user.Email, organization.Name, token, emailMessage)
	err = db.DbCreateInvite(organizationId, inviteeEmail, token, emailMessage, user.Id)
	if err != nil {
		return err
	}

	return nil
}

func AcceptInvite(token string, user common.TrimmedUser) error {
	invite, err := db.DbGetInviteByTokenAndUser(token, user.Email)
	if err != nil {
		return err
	}

	if invite.Accepted {
		return errors.New("invite has already been accepted")
	}

	err = db.DbAcceptInvite(invite.Token, user.Email, user.Id)
	if err != nil {
		return err
	}

	return nil
}

func GetOrganizationInvites(organizationId int64) ([]common.Invite, error) {
	return db.DbGetOrganizationInvites(organizationId)
}
func DbGetInviteByTokenAndUser(token string, userEmail string) (common.Invite, error) {
	return db.DbGetInviteByTokenAndUser(token, userEmail)
}
