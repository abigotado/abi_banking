package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserRole represents user's role in the system
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// UserStatus represents user's status
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusBlocked  UserStatus = "blocked"
	StatusInactive UserStatus = "inactive"
)

// User represents a bank user
type User struct {
	ID          int64      `json:"id"`
	Email       string     `json:"email" validate:"required,email"`
	Username    string     `json:"username" validate:"required,min=3,max=50"`
	Password    string     `json:"-"` // Password hash is never exposed in JSON
	FirstName   string     `json:"first_name" validate:"required"`
	LastName    string     `json:"last_name" validate:"required"`
	PhoneNumber string     `json:"phone_number" validate:"required,e164"`
	Role        UserRole   `json:"role" validate:"required,oneof=user admin"`
	Status      UserStatus `json:"status" validate:"required,oneof=active blocked inactive"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UserResponse represents a safe user response without sensitive data
type UserResponse struct {
	ID          int64      `json:"id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	PhoneNumber string     `json:"phone_number"`
	Role        UserRole   `json:"role"`
	Status      UserStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

// UserSettings represents user's preferences and settings
type UserSettings struct {
	ID                 int64     `json:"id"`
	UserID             int64     `json:"user_id"`
	EmailNotifications bool      `json:"email_notifications"`
	SMSNotifications   bool      `json:"sms_notifications"`
	Language           string    `json:"language" validate:"required,len=2"`
	TimeZone           string    `json:"timezone" validate:"required"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		PhoneNumber: u.PhoneNumber,
		Role:        u.Role,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt,
	}
}
