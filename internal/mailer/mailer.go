package mailer

import (
	"context"
	"fmt"
	"strings"

	"github.com/pmaojo/goploy/internal/config"
	"github.com/pmaojo/goploy/internal/mailer/transport"
	"github.com/jordan-wright/email"
	"github.com/rs/zerolog/log"
)

type Mailer struct {
	Config    config.Mailer
	Transport transport.MailTransporter
}

func New(config config.Mailer, transport transport.MailTransporter) *Mailer {
	return &Mailer{
		Config:    config,
		Transport: transport,
	}
}

func NewWithConfig(cfg config.Mailer, smtpConfig transport.SMTPMailTransportConfig) (*Mailer, error) {
	var mailer *Mailer

	switch config.MailerTransporter(cfg.Transporter) {
	case config.MailerTransporterMock:
		log.Warn().Msg("Initializing mock mailer")
		mailer = New(cfg, transport.NewMock())
	case config.MailerTransporterSMTP:
		mailer = New(cfg, transport.NewSMTP(smtpConfig))
	default:
		return nil, fmt.Errorf("unsupported mail transporter: %s", cfg.Transporter)
	}

	return mailer, nil
}

func (m *Mailer) SendDeploymentNotification(ctx context.Context, to []string, projectName string, status string, output string) error {
	if len(to) == 0 {
		return nil
	}

	if !m.Config.Send {
		log.Warn().Strs("to", to).Msg("Sending has been disabled in mailer config, skipping deployment notification")
		return nil
	}

	subject := fmt.Sprintf("[%s] Deployment %s: %s", strings.ToUpper(status), projectName, status)
	body := fmt.Sprintf("Deployment for project '%s' finished with status: %s.\n\nLogs:\n%s", projectName, status, output)

	mail := email.NewEmail()
	mail.From = m.Config.DefaultSender
	mail.To = to
	mail.Subject = subject
	mail.Text = []byte(body)

	if err := m.Transport.Send(mail); err != nil {
		log.Error().Err(err).Msg("Failed to send deployment notification")
		return fmt.Errorf("failed to send deployment notification: %w", err)
	}

	log.Info().Strs("to", to).Str("project", projectName).Msg("Sent deployment notification")
	return nil
}
