package models

import (
	"time"

	"factory/internal/testutils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Пользователь
type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	KeycloakID string    `gorm:"unique;not null" json:"keycloak_id"`
	Username   string    `gorm:"not null" json:"username"`
	Email      string    `gorm:"not null" json:"email"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`

	Plumbuses []Plumbus `gorm:"foreignKey:UserID" json:"plumbuses,omitempty"`
}

// TableName возвращает имя таблицы для модели User
func (User) TableName() string {
	return "user"
}

// Плюмбус
type Plumbus struct {
	ID            uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID     `gorm:"type:uuid;not null;references:ID" json:"user_id"`
	Name          string        `gorm:"not null" json:"name"`
	Size          string        `gorm:"not null" json:"size"`
	Color         string        `gorm:"not null" json:"color"`
	Shape         string        `gorm:"not null" json:"shape"`
	Weight        string        `gorm:"not null" json:"weight"`
	Wrapping      string        `gorm:"not null" json:"wrapping"`
	Status        PlumbusStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	IsRare        bool          `gorm:"default:false" json:"is_rare"`
	ImagePath     *string       `json:"image_path,omitempty"`
	Signature     *string       `json:"signature,omitempty"`
	SignatureDate *time.Time    `json:"signature_date,omitempty"`
	ErrorMsg      *string       `json:"error_msg,omitempty"`
	CreatedAt     time.Time     `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"not null" json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName возвращает имя таблицы для модели Plumbus
func (Plumbus) TableName() string {
	return "plumbus"
}

// TestUser возвращает структуру User с SQLiteUUID для тестов
func (u *User) TestUser(db *gorm.DB) interface{} {
	if db != nil && db.Name() == "sqlite" {
		return struct {
			ID         testutils.SQLiteUUID `gorm:"primaryKey"`
			KeycloakID string               `gorm:"unique;not null"`
			Username   string               `gorm:"not null"`
			Email      string               `gorm:"not null"`
			CreatedAt  time.Time            `gorm:"not null"`
			UpdatedAt  time.Time            `gorm:"not null"`
		}{
			ID:         testutils.SQLiteUUID(u.ID),
			KeycloakID: u.KeycloakID,
			Username:   u.Username,
			Email:      u.Email,
			CreatedAt:  u.CreatedAt,
			UpdatedAt:  u.UpdatedAt,
		}
	}
	return u
}

// TestPlumbus возвращает структуру Plumbus с SQLiteUUID для тестов
func (p *Plumbus) TestPlumbus(db *gorm.DB) interface{} {
	if db != nil && db.Name() == "sqlite" {
		return struct {
			ID            testutils.SQLiteUUID `gorm:"primaryKey"`
			UserID        testutils.SQLiteUUID `gorm:"not null"`
			Name          string               `gorm:"not null"`
			Size          string               `gorm:"not null"`
			Color         string               `gorm:"not null"`
			Shape         string               `gorm:"not null"`
			Weight        string               `gorm:"not null"`
			Wrapping      string               `gorm:"not null"`
			Status        PlumbusStatus        `gorm:"type:varchar(20);default:'pending'"`
			IsRare        bool                 `gorm:"default:false"`
			ImagePath     *string
			Signature     *string
			SignatureDate *time.Time
			ErrorMsg      *string
			CreatedAt     time.Time `gorm:"not null"`
			UpdatedAt     time.Time `gorm:"not null"`
		}{
			ID:            testutils.SQLiteUUID(p.ID),
			UserID:        testutils.SQLiteUUID(p.UserID),
			Name:          p.Name,
			Size:          p.Size,
			Color:         p.Color,
			Shape:         p.Shape,
			Weight:        p.Weight,
			Wrapping:      p.Wrapping,
			Status:        p.Status,
			IsRare:        p.IsRare,
			ImagePath:     p.ImagePath,
			Signature:     p.Signature,
			SignatureDate: p.SignatureDate,
			ErrorMsg:      p.ErrorMsg,
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
		}
	}
	return p
}

type PlumbusStatus string

const (
	StatusPending    PlumbusStatus = "pending"
	StatusGenerating PlumbusStatus = "generating"
	StatusCompleted  PlumbusStatus = "completed"
	StatusFailed     PlumbusStatus = "failed"
)

// Запрос на генерацию плюмбуса
type PlumbusRequest struct {
	Name     string `json:"name" binding:"required"`
	Size     string `json:"size" binding:"required"`
	Color    string `json:"color" binding:"required"`
	Shape    string `json:"shape" binding:"required"`
	Weight   string `json:"weight" binding:"required"`
	Wrapping string `json:"wrapping" binding:"required"`
}

// Ответ сервиса генерации плюмбуса
type PlumbusGenerationRequest struct {
	Size     string `json:"size"`
	Color    string `json:"color"`
	Shape    string `json:"shape"`
	Weight   string `json:"weight"`
	Wrapping string `json:"wrapping"`
}
