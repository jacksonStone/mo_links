package common

import "time"

type User struct {
	Salt                 string
	HashedPassword       string
	Id                   int64
	Email                string
	ActiveOrganizationId int64
}

type TrimmedUser struct {
	Id                   int64  `json:"id"`
	Email                string `json:"email"`
	ActiveOrganizationId int64  `json:"activeOrganizationId"`
}

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

type UserDetails struct {
	Id                   int64
	Email                string
	ActiveOrganizationId int64
	Memberships          []OrganizationMember
	MoLinks              []MoLink
}
type MoLink struct {
	Id             int64
	Name           string
	Url            string
	OrganizationId int64
	CreatedAt      time.Time
	Views          int64
}
type MembershipInvite struct {
	Id              int64     `json:"id"`
	OrganizationId  int64     `json:"organizationId"`
	InviteeEmail    string    `json:"inviteeEmail"`
	Token           string    `json:"token"`
	EmailMessage    string    `json:"emailMessage"`
	SentAt          time.Time `json:"sentAt"`
	CreatedByUserId int64     `json:"createdByUserId"`
	AcceptedAt      time.Time `json:"acceptedAt"`
	InviteeId       int64     `json:"inviteeId"`
	Accepted        bool      `json:"accepted"`
}

const (
	RoleAdmin  = "Admin"
	RoleOwner  = "Owner"
	RoleMember = "Member"
)

const (
	OrgNamePersonal = "Personal"
)
