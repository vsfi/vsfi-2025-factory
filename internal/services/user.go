package services

import (
	"factory/internal/models"
	"factory/internal/testutils"

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

// SQLiteUser структура для работы с SQLite
type SQLiteUser struct {
	ID         testutils.SQLiteUUID `gorm:"primaryKey"`
	KeycloakID string               `gorm:"unique;not null"`
	Username   string               `gorm:"not null"`
	Email      string               `gorm:"not null"`
	CreatedAt  time.Time            `gorm:"not null"`
	UpdatedAt  time.Time            `gorm:"not null"`
}

func (SQLiteUser) TableName() string {
	return "user"
}

// SQLitePlumbus структура для работы с SQLite
type SQLitePlumbus struct {
	ID            testutils.SQLiteUUID `gorm:"primaryKey"`
	UserID        testutils.SQLiteUUID `gorm:"not null"`
	Name          string               `gorm:"not null"`
	Size          string               `gorm:"not null"`
	Color         string               `gorm:"not null"`
	Shape         string               `gorm:"not null"`
	Weight        string               `gorm:"not null"`
	Wrapping      string               `gorm:"not null"`
	Status        models.PlumbusStatus `gorm:"type:varchar(20);default:'pending'"`
	IsRare        bool                 `gorm:"default:false"`
	ImagePath     *string
	Signature     *string
	SignatureDate *time.Time
	ErrorMsg      *string
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (SQLitePlumbus) TableName() string {
	return "plumbus"
}

func (s *UserService) GetOrCreateUser(keycloakID, username, email string) (*models.User, error) {
	if s.db.Name() == "sqlite" {
		var sqliteUser SQLiteUser

		// Ищем пользователя по Keycloak ID
		err := s.db.Where("keycloak_id = ?", keycloakID).First(&sqliteUser).Error
		if err == nil {
			return &models.User{
				ID:         uuid.UUID(sqliteUser.ID),
				KeycloakID: sqliteUser.KeycloakID,
				Username:   sqliteUser.Username,
				Email:      sqliteUser.Email,
				CreatedAt:  sqliteUser.CreatedAt,
				UpdatedAt:  sqliteUser.UpdatedAt,
			}, nil
		}

		// Если пользователь не найден, создаем нового
		if err == gorm.ErrRecordNotFound {
			newID := uuid.New()
			sqliteUser = SQLiteUser{
				ID:         testutils.SQLiteUUID(newID),
				KeycloakID: keycloakID,
				Username:   username,
				Email:      email,
			}

			if err := s.db.Create(&sqliteUser).Error; err != nil {
				return nil, err
			}

			return &models.User{
				ID:         newID,
				KeycloakID: sqliteUser.KeycloakID,
				Username:   sqliteUser.Username,
				Email:      sqliteUser.Email,
				CreatedAt:  sqliteUser.CreatedAt,
				UpdatedAt:  sqliteUser.UpdatedAt,
			}, nil
		}

		return nil, err
	}

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

	if s.db.Name() == "sqlite" {
		newID := uuid.New()
		sqlitePlumbus := SQLitePlumbus{
			ID:       testutils.SQLiteUUID(newID),
			UserID:   testutils.SQLiteUUID(userID),
			Name:     req.Name,
			Size:     req.Size,
			Color:    req.Color,
			Shape:    req.Shape,
			Weight:   req.Weight,
			Wrapping: req.Wrapping,
			Status:   models.StatusPending,
			IsRare:   isRare,
		}

		if err := s.db.Create(&sqlitePlumbus).Error; err != nil {
			return nil, err
		}

		plumbus := &models.Plumbus{
			ID:        newID,
			UserID:    userID,
			Name:      sqlitePlumbus.Name,
			Size:      sqlitePlumbus.Size,
			Color:     sqlitePlumbus.Color,
			Shape:     sqlitePlumbus.Shape,
			Weight:    sqlitePlumbus.Weight,
			Wrapping:  sqlitePlumbus.Wrapping,
			Status:    sqlitePlumbus.Status,
			IsRare:    sqlitePlumbus.IsRare,
			CreatedAt: sqlitePlumbus.CreatedAt,
			UpdatedAt: sqlitePlumbus.UpdatedAt,
		}

		// Логируем создание редкого плюмбуса
		if isRare {
			log.Printf("🌟 RARE PLUMBUS CREATED! ID: %s, Name: %s, User: %s", plumbus.ID, plumbus.Name, userID)
		}

		return plumbus, nil
	}

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

	if s.db.Name() == "sqlite" {
		return s.db.Model(&SQLitePlumbus{}).Where("id = ?", testutils.SQLiteUUID(id)).Updates(updates).Error
	}

	return s.db.Model(&models.Plumbus{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) GetPlumbus(id uuid.UUID) (*models.Plumbus, error) {
	if s.db.Name() == "sqlite" {
		var sqlitePlumbus SQLitePlumbus
		err := s.db.First(&sqlitePlumbus, "id = ?", testutils.SQLiteUUID(id)).Error
		if err != nil {
			return nil, err
		}

		return &models.Plumbus{
			ID:            uuid.UUID(sqlitePlumbus.ID),
			UserID:        uuid.UUID(sqlitePlumbus.UserID),
			Name:          sqlitePlumbus.Name,
			Size:          sqlitePlumbus.Size,
			Color:         sqlitePlumbus.Color,
			Shape:         sqlitePlumbus.Shape,
			Weight:        sqlitePlumbus.Weight,
			Wrapping:      sqlitePlumbus.Wrapping,
			Status:        sqlitePlumbus.Status,
			IsRare:        sqlitePlumbus.IsRare,
			ImagePath:     sqlitePlumbus.ImagePath,
			Signature:     sqlitePlumbus.Signature,
			SignatureDate: sqlitePlumbus.SignatureDate,
			ErrorMsg:      sqlitePlumbus.ErrorMsg,
			CreatedAt:     sqlitePlumbus.CreatedAt,
			UpdatedAt:     sqlitePlumbus.UpdatedAt,
		}, nil
	}

	var plumbus models.Plumbus
	err := s.db.First(&plumbus, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &plumbus, nil
}

func (s *UserService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	if s.db.Name() == "sqlite" {
		var sqliteUser SQLiteUser
		err := s.db.First(&sqliteUser, "id = ?", testutils.SQLiteUUID(userID)).Error
		if err != nil {
			return nil, err
		}

		return &models.User{
			ID:         uuid.UUID(sqliteUser.ID),
			KeycloakID: sqliteUser.KeycloakID,
			Username:   sqliteUser.Username,
			Email:      sqliteUser.Email,
			CreatedAt:  sqliteUser.CreatedAt,
			UpdatedAt:  sqliteUser.UpdatedAt,
		}, nil
	}

	var user models.User
	err := s.db.First(&user, "id = ?", userID).Error
	return &user, err
}

func (s *UserService) GetUserPlumbuses(userID uuid.UUID) ([]models.Plumbus, error) {
	if s.db.Name() == "sqlite" {
		var sqlitePlumbuses []SQLitePlumbus
		err := s.db.Where("user_id = ?", testutils.SQLiteUUID(userID)).Order("created_at desc").Find(&sqlitePlumbuses).Error
		if err != nil {
			return nil, err
		}

		plumbuses := make([]models.Plumbus, len(sqlitePlumbuses))
		for i, sp := range sqlitePlumbuses {
			plumbuses[i] = models.Plumbus{
				ID:            uuid.UUID(sp.ID),
				UserID:        uuid.UUID(sp.UserID),
				Name:          sp.Name,
				Size:          sp.Size,
				Color:         sp.Color,
				Shape:         sp.Shape,
				Weight:        sp.Weight,
				Wrapping:      sp.Wrapping,
				Status:        sp.Status,
				IsRare:        sp.IsRare,
				ImagePath:     sp.ImagePath,
				Signature:     sp.Signature,
				SignatureDate: sp.SignatureDate,
				ErrorMsg:      sp.ErrorMsg,
				CreatedAt:     sp.CreatedAt,
				UpdatedAt:     sp.UpdatedAt,
			}
		}
		return plumbuses, nil
	}

	var plumbuses []models.Plumbus
	err := s.db.Where("user_id = ?", userID).Order("created_at desc").Find(&plumbuses).Error
	return plumbuses, err
}
