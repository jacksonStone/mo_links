package utilities

import (
	"fmt"
	"net/url"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func getRootUrl() string {
	envType := os.Getenv("NODE_ENV")
	if envType == "production" {
		return "https://molinks.me"
	}
	return "http://localhost:3003"
}

func SendInviteEmail(inviteeEmail, inviterEmail, organizationName, verificationToken, emailMessage string) error {
	encodedEmail := url.QueryEscape(inviteeEmail)

	return sendEmail(
		inviteeEmail,
		"You've been invited to join a Molinks.me Organization!",
		"admin@molinks.me",
		"admin@molinks.me",
		fmt.Sprintf(`
You've been invited by %s to join the MoLinks organization: %s
Their Note: %s
Click the link below to accept the invite:
%s/____reserved/api/accept_invite?token=%s&email=%s`,
			inviterEmail, organizationName, emailMessage, getRootUrl(), verificationToken, encodedEmail),

		fmt.Sprintf(`
<strong>You've been invited by %s to join the MoLinks organization: %s</strong>
<br>
Their Note: %s
<br>
Click the link below to accept the invite:
<br>
<a href="%s/____reserved/api/accept_invite?token=%s&email=%s">Accept Invite</a>`,
			inviterEmail, organizationName, emailMessage, getRootUrl(), verificationToken, encodedEmail))
}

func SendSignupValidationEmail(userEmail, verificationToken string) error {
	encodedEmail := url.QueryEscape(userEmail)

	return sendEmail(userEmail, "Welcome to MoLinks!", "admin@molinks.me", "admin@molinks.me",
		fmt.Sprintf(`
Welcome to MoLinks!
Copy and paste the following link into your browser to verify your email:
%s/____reserved/api/verify_email?token=%s&email=%s`,
			getRootUrl(), verificationToken, encodedEmail),

		fmt.Sprintf(`
<strong>Welcome to MoLinks!</strong>
<br>
Click the link below to verify your email:
<br>
<a href="%s/____reserved/api/verify_email?token=%s&email=%s">Verify Email</a>`,
			getRootUrl(), verificationToken, encodedEmail))
}

// Use sendgrid to send emails
func sendEmail(to, subject, fromName, fromEmail, plainTextContent, htmlContent string) error {
	formattedFrom := mail.NewEmail(fromName, fromEmail)
	formattedTo := mail.NewEmail(to, to)
	// from := mail.NewEmail("Example User", "test@example.com")
	// subject := "Sending with Twilio SendGrid is Fun"
	// to := mail.NewEmail("Example User", "test@example.com")
	// Plain textContent is a fallback for htmlContent
	message := mail.NewSingleEmail(formattedFrom, subject, formattedTo, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("MO_LINKS_SENDGRID_API_KEY"))
	_, err := client.Send(message)
	return err
}
