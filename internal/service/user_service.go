package service

import (
	"errors"
	"time"

	"github.com/Abigotado/abi_banking/internal/middleware"
	"github.com/Abigotado/abi_banking/internal/models"
	"github.com/Abigotado/abi_banking/internal/repository"
	"github.com/sirupsen/logrus"
)

type UserService struct {
	userRepo *repository.UserRepository
	logger   *logrus.Logger
}

func NewUserService(logger *logrus.Logger) *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
		logger:   logger,
	}
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (s *UserService) Register(req *RegisterRequest) error {
	// Check if email exists
	emailExists, err := s.userRepo.CheckEmailExists(req.Email)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check email existence")
		return errors.New("internal server error")
	}
	if emailExists {
		return errors.New("email already exists")
	}

	// Check if username exists
	usernameExists, err := s.userRepo.CheckUsernameExists(req.Username)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check username existence")
		return errors.New("internal server error")
	}
	if usernameExists {
		return errors.New("username already exists")
	}

	// Create user
	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		s.logger.WithError(err).Error("Failed to hash password")
		return errors.New("internal server error")
	}

	// Save user
	if err := s.userRepo.Create(user); err != nil {
		s.logger.WithError(err).Error("Failed to create user")
		return errors.New("internal server error")
	}

	return nil
}

func (s *UserService) Login(req *LoginRequest) (*LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user by email")
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate token")
		return nil, errors.New("internal server error")
	}

	return &LoginResponse{
		Token: token,
	}, nil
}

func (s *UserService) GetUserByID(userID int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user by ID")
		return nil, errors.New("user not found")
	}

	// Clear sensitive data
	user.Password = ""

	return user, nil
}
