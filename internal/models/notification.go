package models

import "time"

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending  NotificationStatus = "pending"
	NotificationStatusSent     NotificationStatus = "sent"
	NotificationStatusFailed   NotificationStatus = "failed"
	NotificationStatusCanceled NotificationStatus = "canceled"
)

// NotificationPriority represents the priority of a notification
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityNormal NotificationPriority = "normal"
	PriorityHigh   NotificationPriority = "high"
)

// Notification represents a notification to be sent to a user
type Notification struct {
	ID         int64                `json:"id"`
	UserID     int64                `json:"user_id" validate:"required"`
	Type       NotificationType     `json:"type" validate:"required,oneof=email sms"`
	Priority   NotificationPriority `json:"priority" validate:"required,oneof=low normal high"`
	Status     NotificationStatus   `json:"status" validate:"required"`
	Subject    string               `json:"subject" validate:"required"`
	Content    string               `json:"content" validate:"required"`
	Recipient  string               `json:"recipient" validate:"required"` // email or phone number
	SentAt     *time.Time           `json:"sent_at,omitempty"`
	Error      string               `json:"error,omitempty"`
	RetryCount int                  `json:"retry_count"`
	MaxRetries int                  `json:"max_retries"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

// NotificationTemplate represents a template for notifications
type NotificationTemplate struct {
	ID        int64            `json:"id"`
	Name      string           `json:"name" validate:"required"`
	Type      NotificationType `json:"type" validate:"required,oneof=email sms"`
	Subject   string           `json:"subject"`
	Content   string           `json:"content" validate:"required"`
	Variables []string         `json:"variables"` // List of variables used in template
	IsActive  bool             `json:"is_active"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID     int64                `json:"user_id" validate:"required"`
	Type       NotificationType     `json:"type" validate:"required,oneof=email sms"`
	Priority   NotificationPriority `json:"priority" validate:"required,oneof=low normal high"`
	Subject    string               `json:"subject" validate:"required"`
	Content    string               `json:"content" validate:"required"`
	MaxRetries int                  `json:"max_retries" validate:"required,min=0,max=5"`
}

// NotificationResponse represents a notification response
type NotificationResponse struct {
	ID        int64              `json:"id"`
	Type      NotificationType   `json:"type"`
	Status    NotificationStatus `json:"status"`
	Subject   string             `json:"subject"`
	SentAt    *time.Time         `json:"sent_at,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
}
