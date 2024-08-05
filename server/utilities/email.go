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

func SendInviteEmail(inviteeEmail, inviterEmail, organizationName, verificationToken string) error {
	encodedEmail := url.QueryEscape(inviteeEmail)

	return sendEmail(
		inviteeEmail,
		"You've been invited to join a Molinks.me Organization!",
		"admin@molinks.me",
		"admin@molinks.me",
		fmt.Sprintf(`
You've been invited by %s to join the MoLinks organization: %s
Click the link below to accept the invite:
%s/___reserved/api/accept-invite?token=%s&email=%s`,
			inviterEmail, organizationName, getRootUrl(), verificationToken, encodedEmail),

		fmt.Sprintf(`
<strong>You've been invited by %s to join the MoLinks organization: %s</strong>
<br>
Click the link below to accept the invite:
<br>
<a href="%s/___reserved/api/accept-invite?token=%s&email=%s">Accept Invite</a>`,
			inviterEmail, organizationName, getRootUrl(), verificationToken, encodedEmail))
}

func SendSignupValidationEmail(userEmail, verificationToken string) error {
	return sendEmail(userEmail, "Welcome to MoLinks!", "admin@molinks.me", "admin@molinks.me",
		fmt.Sprintf(`
Welcome to MoLinks!
Copy and paste the following link into your browser to verify your email:
%s/___reserved/api/verify-email?token=%s`,
			getRootUrl(), verificationToken),

		fmt.Sprintf(`
<strong>Welcome to MoLinks!</strong>
<br>
Click the link below to verify your email:
<br>
<a href="%s/___reserved/api/verify-email?token=%s">Verify Email</a>`,
			getRootUrl(), verificationToken))
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
	response, err := client.Send(message)
	if err != nil {
		fmt.Println(err)
		return err
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
	return nil
}
