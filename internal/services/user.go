package services

import (
	"factory/internal/models"

	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetOrCreateUser(keycloakID, username, email string) (*models.User, error) {
	var user models.User

	// Ищем пользователя по Keycloak ID
	err := s.db.Where("keycloak_id = ?", keycloakID).First(&user).Error
	if err == nil {
		return &user, nil
	}

	// Если пользователь не найден, создаем нового
	if err == gorm.ErrRecordNotFound {
		user = models.User{
			KeycloakID: keycloakID,
			Username:   username,
			Email:      email,
		}

		if err := s.db.Create(&user).Error; err != nil {
			return nil, err
		}

		return &user, nil
	}

	return nil, err
}

func (s *UserService) CreatePlumbus(userID uuid.UUID, req models.PlumbusRequest) (*models.Plumbus, error) {
	// Генерируем случайность для редкого плюмбуса (5% шанс)
	rand.Seed(time.Now().UnixNano())
	isRare := rand.Float64() < 0.05

	plumbus := models.Plumbus{
		UserID:   userID,
		Name:     req.Name,
		Size:     req.Size,
		Color:    req.Color,
		Shape:    req.Shape,
		Weight:   req.Weight,
		Wrapping: req.Wrapping,
		Status:   models.StatusPending,
		IsRare:   isRare,
	}

	if err := s.db.Create(&plumbus).Error; err != nil {
		return nil, err
	}

	// Логируем создание редкого плюмбуса
	if isRare {
		log.Printf("🌟 RARE PLUMBUS CREATED! ID: %s, Name: %s, User: %s", plumbus.ID, plumbus.Name, userID)
	}

	return &plumbus, nil
}

func (s *UserService) UpdatePlumbusStatus(id uuid.UUID, status models.PlumbusStatus, imagePath *string, errorMsg *string, signature *string, signatureDate *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if imagePath != nil {
		updates["image_path"] = *imagePath
	}

	if errorMsg != nil {
		updates["error_msg"] = *errorMsg
	}

	if signature != nil {
		updates["signature"] = *signature
	}

	if signatureDate != nil {
		updates["signature_date"] = *signatureDate
	}

	return s.db.Model(&models.Plumbus{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) GetPlumbus(id uuid.UUID) (*models.Plumbus, error) {
	var plumbus models.Plumbus
	err := s.db.First(&plumbus, "id = ?", id).Error
	return &plumbus, err
}

func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, "id = ?", userID).Error
	return &user, err
}

func (s *UserService) GetUserPlumbuses(userID uuid.UUID) ([]models.Plumbus, error) {
	var plumbuses []models.Plumbus
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&plumbuses).Error
	return plumbuses, err
}
