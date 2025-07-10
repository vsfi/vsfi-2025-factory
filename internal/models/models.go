package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Пользователь
type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	KeycloakID string    `gorm:"unique;not null" json:"keycloak_id"`
	Username   string    `gorm:"not null" json:"username"`
	Email      string    `gorm:"not null" json:"email"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Plumbuses []Plumbus `gorm:"foreignKey:UserID" json:"plumbuses,omitempty"`
}

// Плюмбус
type Plumbus struct {
	ID            uuid.UUID     `gorm:"type:uuid;primary_key;" json:"id"`
	UserID        uuid.UUID     `gorm:"type:uuid;not null" json:"user_id"`
	Name          string        `gorm:"not null" json:"name"`
	Size          string        `gorm:"not null" json:"size"`
	Color         string        `gorm:"not null" json:"color"`
	Shape         string        `gorm:"not null" json:"shape"`
	Weight        string        `gorm:"not null" json:"weight"`
	Wrapping      string        `gorm:"not null" json:"wrapping"`
	Status        PlumbusStatus `gorm:"default:'pending'" json:"status"`
	IsRare        bool          `gorm:"default:false" json:"is_rare"`
	ImagePath     *string       `json:"image_path,omitempty"`
	Signature     *string       `json:"signature,omitempty"`
	SignatureDate *time.Time    `json:"signature_date,omitempty"`
	ErrorMsg      *string       `json:"error_msg,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
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

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}

func (p *Plumbus) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}
