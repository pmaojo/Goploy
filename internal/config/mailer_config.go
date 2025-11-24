package config

import (
	"strings"
)

type MailerTransporter string

func (e MailerTransporter) String() string {
	return string(e)
}

const (
	MailerTransporterMock MailerTransporter = "mock"
	MailerTransporterSMTP MailerTransporter = "smtp"
)

type Mailer struct {
	DefaultSender               string
	Send                        bool
	WebTemplatesEmailBaseDirAbs string
	Transporter                 string
}

func (m Mailer) TransporterEnum() MailerTransporter {
	return MailerTransporter(m.Transporter)
}

func (m Mailer) ParseTemplateNames() []string {
	// Not used anymore as we embed/hardcode templates
	return []string{}
}

func (m Mailer) IsValidTransporter() bool {
	switch m.TransporterEnum() {
	case MailerTransporterMock, MailerTransporterSMTP:
		return true
	default:
		return false
	}
}

func (m Mailer) IsMock() bool {
	return m.TransporterEnum() == MailerTransporterMock
}

func (m Mailer) IsSMTP() bool {
	return m.TransporterEnum() == MailerTransporterSMTP
}

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
