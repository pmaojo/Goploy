package config

import (
	"strings"
)

// MailerTransporter represents the type of email transport mechanism.
type MailerTransporter string

// String returns the string representation of the MailerTransporter.
func (e MailerTransporter) String() string {
	return string(e)
}

const (
	// MailerTransporterMock indicates a mock mailer for testing.
	MailerTransporterMock MailerTransporter = "mock"
	// MailerTransporterSMTP indicates an SMTP mailer for real email delivery.
	MailerTransporterSMTP MailerTransporter = "smtp"
)

// Mailer holds configuration settings for the email service.
type Mailer struct {
	// DefaultSender is the email address used as the sender.
	DefaultSender string
	// Send indicates whether emails should actually be sent.
	Send bool
	// WebTemplatesEmailBaseDirAbs is the absolute path to the email templates directory (deprecated/unused if embedded).
	WebTemplatesEmailBaseDirAbs string
	// Transporter is the string representation of the transport mechanism (e.g., "smtp", "mock").
	Transporter string
}

// TransporterEnum converts the string Transporter field to a MailerTransporter enum.
//
// Returns:
//   The MailerTransporter enum value.
func (m Mailer) TransporterEnum() MailerTransporter {
	return MailerTransporter(m.Transporter)
}

// ParseTemplateNames returns a list of template names.
// Note: This method currently returns an empty list as templates are embedded or hardcoded.
//
// Returns:
//   An empty slice of strings.
func (m Mailer) ParseTemplateNames() []string {
	// Not used anymore as we embed/hardcode templates
	return []string{}
}

// IsValidTransporter checks if the configured transporter is a valid known type.
//
// Returns:
//   True if the transporter is either Mock or SMTP, false otherwise.
func (m Mailer) IsValidTransporter() bool {
	switch m.TransporterEnum() {
	case MailerTransporterMock, MailerTransporterSMTP:
		return true
	default:
		return false
	}
}

// IsMock checks if the transporter is set to Mock.
//
// Returns:
//   True if the transporter is Mock.
func (m Mailer) IsMock() bool {
	return m.TransporterEnum() == MailerTransporterMock
}

// IsSMTP checks if the transporter is set to SMTP.
//
// Returns:
//   True if the transporter is SMTP.
func (m Mailer) IsSMTP() bool {
	return m.TransporterEnum() == MailerTransporterSMTP
}

// MailerTransporterFromString converts a string to a MailerTransporter enum value.
// It defaults to MailerTransporterMock if the string is not recognized.
//
// Parameters:
//   - val: The string representation of the transporter.
//
// Returns:
//   The corresponding MailerTransporter enum.
func MailerTransporterFromString(val string) MailerTransporter {
	switch strings.ToLower(val) {
	case "mock":
		return MailerTransporterMock
	case "smtp":
		return MailerTransporterSMTP
	default:
		return MailerTransporterMock
	}
}
