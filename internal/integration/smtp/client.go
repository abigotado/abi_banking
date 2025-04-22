package smtp

import (
	"crypto/tls"
	"fmt"

	"github.com/Abigotado/abi_banking/internal/config"
	"github.com/Abigotado/abi_banking/internal/models"
	"gopkg.in/mail.v2"
)

// Client represents an SMTP client
type Client struct {
	config *config.SMTPConfig
	dialer *mail.Dialer
}

// NewClient creates a new SMTP client
func NewClient(config *config.SMTPConfig) *Client {
	dialer := mail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	dialer.TLSConfig = &tls.Config{
		ServerName:         config.Host,
		InsecureSkipVerify: false,
	}

	return &Client{
		config: config,
		dialer: dialer,
	}
}

// SendEmail sends an email using the configured SMTP server
func (c *Client) SendEmail(notification *models.Notification) error {
	m := mail.NewMessage()

	// Set headers
	m.SetHeader("From", c.config.From)
	m.SetHeader("To", notification.Recipient)
	m.SetHeader("Subject", notification.Subject)
	m.SetBody("text/html", notification.Content)

	// Send email
	if err := c.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendBulkEmails sends multiple emails in batch
func (c *Client) SendBulkEmails(notifications []*models.Notification) []error {
	errors := make([]error, 0)

	// Create a connection
	s, err := c.dialer.Dial()
	if err != nil {
		return []error{fmt.Errorf("failed to connect to SMTP server: %w", err)}
	}
	defer s.Close()

	// Send emails using the same connection
	for _, notification := range notifications {
		m := mail.NewMessage()
		m.SetHeader("From", c.config.From)
		m.SetHeader("To", notification.Recipient)
		m.SetHeader("Subject", notification.Subject)
		m.SetBody("text/html", notification.Content)

		if err := mail.Send(s, m); err != nil {
			errors = append(errors, fmt.Errorf("failed to send email to %s: %w", notification.Recipient, err))
		}
		m.Reset()
	}

	return errors
}

// SendTemplate sends an email using a template
func (c *Client) SendTemplate(template *models.NotificationTemplate, recipient string, data map[string]interface{}) error {
	// Create a new message
	m := mail.NewMessage()

	// Set headers
	m.SetHeader("From", c.config.From)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", template.Subject)

	// TODO: Implement template rendering with data
	// For now, just use the template content as is
	m.SetBody("text/html", template.Content)

	// Send email
	if err := c.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send template email: %w", err)
	}

	return nil
}
