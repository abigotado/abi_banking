package service

import (
	"errors"
	"time"

	"github.com/Abigotado/abi_banking/internal/models"
	"github.com/Abigotado/abi_banking/internal/repository"
	"github.com/sirupsen/logrus"
)

// CardService handles business logic for card operations
type CardService struct {
	cardRepo    *repository.CardRepository
	accountRepo *repository.AccountRepository
	logger      *logrus.Logger
}

// NewCardService creates a new CardService instance
func NewCardService(
	cardRepo *repository.CardRepository,
	accountRepo *repository.AccountRepository,
	logger *logrus.Logger,
) *CardService {
	return &CardService{
		cardRepo:    cardRepo,
		accountRepo: accountRepo,
		logger:      logger,
	}
}

// CreateCard creates a new card for a user's account
func (s *CardService) CreateCard(userID int64, req *models.CreateCardRequest) (*models.Card, error) {
	// Validate account ownership
	account, err := s.accountRepo.GetByID(req.AccountID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get account")
		return nil, err
	}
	if account == nil {
		return nil, errors.New("account not found")
	}
	if account.UserID != userID {
		return nil, errors.New("unauthorized: account does not belong to user")
	}

	// Generate card number and expiry date
	cardNumber := generateCardNumber()
	expiryDate := time.Now().AddDate(5, 0, 0).Format("01/06")
	cvv := generateCVV()

	card := &models.Card{
		UserID:     userID,
		AccountID:  req.AccountID,
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		CVV:        cvv,
		CardType:   req.CardType,
		Status:     models.CardStatusActive,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.cardRepo.Create(card); err != nil {
		s.logger.WithError(err).Error("Failed to create card")
		return nil, err
	}

	return card, nil
}

// GetCard retrieves a card by its ID
func (s *CardService) GetCard(userID int64, cardID int64) (*models.Card, error) {
	card, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get card")
		return nil, err
	}
	if card == nil {
		return nil, errors.New("card not found")
	}
	if card.UserID != userID {
		return nil, errors.New("unauthorized: card does not belong to user")
	}

	return card, nil
}

// GetUserCards retrieves all cards for a user
func (s *CardService) GetUserCards(userID int64) ([]*models.Card, error) {
	cards, err := s.cardRepo.GetByUserID(userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user cards")
		return nil, err
	}

	return cards, nil
}

// BlockCard blocks a card
func (s *CardService) BlockCard(userID int64, cardID int64) error {
	card, err := s.GetCard(userID, cardID)
	if err != nil {
		return err
	}

	if card.Status == models.CardStatusBlocked {
		return errors.New("card is already blocked")
	}

	if err := s.cardRepo.UpdateStatus(cardID, models.CardStatusBlocked); err != nil {
		s.logger.WithError(err).Error("Failed to block card")
		return err
	}

	return nil
}

// UnblockCard unblocks a card
func (s *CardService) UnblockCard(userID int64, cardID int64) error {
	card, err := s.GetCard(userID, cardID)
	if err != nil {
		return err
	}

	if card.Status == models.CardStatusActive {
		return errors.New("card is already active")
	}

	if err := s.cardRepo.UpdateStatus(cardID, models.CardStatusActive); err != nil {
		s.logger.WithError(err).Error("Failed to unblock card")
		return err
	}

	return nil
}

// DeleteCard deletes a card
func (s *CardService) DeleteCard(userID int64, cardID int64) error {
	card, err := s.GetCard(userID, cardID)
	if err != nil {
		return err
	}

	if card.Status != models.CardStatusBlocked {
		return errors.New("card must be blocked before deletion")
	}

	if err := s.cardRepo.Delete(cardID); err != nil {
		s.logger.WithError(err).Error("Failed to delete card")
		return err
	}

	return nil
}

// Helper functions
func generateCardNumber() string {
	// TODO: Implement proper card number generation with Luhn algorithm
	return "4111111111111111"
}

func generateCVV() string {
	// TODO: Implement proper CVV generation
	return "123"
}
