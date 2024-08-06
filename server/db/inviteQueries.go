package db

import (
	"database/sql"
	"mo_links/common"
	"time"
)

func initializeInviteQueries() {
	createInviteStmt()
	getInviteByTokenAndUserStmt()
	acceptInviteStmt()
	getOrganizationInvitesStmt()
}

func createInviteStmt() *sql.Stmt {
	return getQuery(`
	INSERT INTO mo_links_membership_invites (organization_id, invitee_email, token, email_message, created_by_user_id)
	VALUES (?, ?, ?, ?, ?)`)
}

func DbCreateInvite(organizationId int64, inviteeEmail string, token string, emailMessage string, createdByUserId int64) error {
	_, err := createInviteStmt().Exec(organizationId, inviteeEmail, token, emailMessage, createdByUserId)
	return err
}

func getInviteByTokenAndUserStmt() *sql.Stmt {
	return getQuery(`
	SELECT id, organization_id, invitee_email, token, email_message, sent_at, created_by_user_id, invitee_id, accepted
	FROM mo_links_membership_invites
	WHERE token = ? AND invitee_email = ?`)
}

func DbGetInviteByTokenAndUser(token string, inviteeEmail string) (common.Invite, error) {
	var invite common.Invite
	var acceptedAt sql.NullTime
	err := getInviteByTokenAndUserStmt().QueryRow(token, inviteeEmail).Scan(
		&invite.Id, &invite.OrganizationId, &invite.InviteeEmail, &invite.Token,
		&invite.EmailMessage, &invite.SentAt, &invite.CreatedByUserId,
		&acceptedAt, &invite.InviteeId, &invite.Accepted,
	)
	if err != nil {
		return common.Invite{}, err
	}
	if acceptedAt.Valid {
		invite.AcceptedAt = acceptedAt.Time
	}
	return invite, nil
}

func acceptInviteStmt() *sql.Stmt {
	return getQuery(`
	UPDATE mo_links_membership_invites
	SET accepted = TRUE, accepted_at = ?, invitee_id = ?
	WHERE id = ?`)
}

func DbAcceptInvite(token string, inviteeEmail string, inviteeId int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // This will be a no-op if the transaction is committed
	var invite common.Invite
	err = tx.Stmt(getInviteByTokenAndUserStmt()).QueryRow(token, inviteeEmail).Scan(
		&invite.Id, &invite.OrganizationId, &invite.InviteeEmail, &invite.Token,
		&invite.EmailMessage, &invite.SentAt, &invite.CreatedByUserId, &invite.InviteeId, &invite.Accepted,
	)
	if err != nil {
		return err
	}
	_, err = tx.Stmt(acceptInviteStmt()).Exec(time.Now(), inviteeId, invite.Id)
	if err != nil {
		return err
	}
	_, err = tx.Stmt(assignMemberToOrganizationStmt()).Exec(inviteeId, common.RoleMember, invite.OrganizationId)
	if err != nil {
		return err
	}
	_, err = tx.Stmt(setUserActiveOrganizationStmt()).Exec(invite.OrganizationId, inviteeId)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

func getOrganizationInvitesStmt() *sql.Stmt {
	return getQuery(`
	SELECT id, invitee_email, email_message, sent_at, created_by_user_id, accepted_at, invitee_id, accepted
	FROM mo_links_membership_invites
	WHERE organization_id = ?`)
}

func DbGetOrganizationInvites(organizationId int64) ([]common.Invite, error) {
	rows, err := getOrganizationInvitesStmt().Query(organizationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []common.Invite
	for rows.Next() {
		var invite common.Invite
		var acceptedAt sql.NullTime
		err = rows.Scan(
			&invite.Id, &invite.InviteeEmail, &invite.EmailMessage,
			&invite.SentAt, &invite.CreatedByUserId, &acceptedAt, &invite.InviteeId, &invite.Accepted,
		)
		if err != nil {
			return nil, err
		}
		invite.OrganizationId = organizationId
		if acceptedAt.Valid {
			invite.AcceptedAt = acceptedAt.Time
		}
		invites = append(invites, invite)
	}
	return invites, nil
}
