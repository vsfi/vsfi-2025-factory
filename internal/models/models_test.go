package models

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestUser_BeforeCreate(t *testing.T) {
	user := &User{
		KeycloakID: "test-keycloak-id",
		Username:   "testuser",
		Email:      "test@example.com",
	}

	// Вызываем BeforeCreate
	err := user.BeforeCreate(&gorm.DB{})

	// Проверяем что ошибки нет
	if err != nil {
		t.Errorf("BeforeCreate() error = %v, want nil", err)
	}

	// Проверяем что ID был установлен
	if user.ID == uuid.Nil {
		t.Error("BeforeCreate() did not set ID")
	}

	// Проверяем что ID валидный UUID
	if _, err := uuid.Parse(user.ID.String()); err != nil {
		t.Errorf("BeforeCreate() set invalid UUID: %v", err)
	}
}

func TestPlumbus_BeforeCreate(t *testing.T) {
	plumbus := &Plumbus{
		UserID:   uuid.New(),
		Name:     "Test Plumbus",
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
		Status:   StatusPending,
	}

	// Вызываем BeforeCreate
	err := plumbus.BeforeCreate(&gorm.DB{})

	// Проверяем что ошибки нет
	if err != nil {
		t.Errorf("BeforeCreate() error = %v, want nil", err)
	}

	// Проверяем что ID был установлен
	if plumbus.ID == uuid.Nil {
		t.Error("BeforeCreate() did not set ID")
	}

	// Проверяем что ID валидный UUID
	if _, err := uuid.Parse(plumbus.ID.String()); err != nil {
		t.Errorf("BeforeCreate() set invalid UUID: %v", err)
	}
}

func TestPlumbusStatus_Constants(t *testing.T) {
	tests := []struct {
		status   PlumbusStatus
		expected string
	}{
		{StatusPending, "pending"},
		{StatusGenerating, "generating"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("PlumbusStatus constant %s = %v, want %v", tt.expected, tt.status, tt.expected)
			}
		})
	}
}

func TestPlumbusRequest_Validation(t *testing.T) {
	validRequest := PlumbusRequest{
		Name:     "Test Plumbus",
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	// Проверяем что все поля заполнены
	if validRequest.Name == "" {
		t.Error("PlumbusRequest.Name should not be empty")
	}
	if validRequest.Size == "" {
		t.Error("PlumbusRequest.Size should not be empty")
	}
	if validRequest.Color == "" {
		t.Error("PlumbusRequest.Color should not be empty")
	}
	if validRequest.Shape == "" {
		t.Error("PlumbusRequest.Shape should not be empty")
	}
	if validRequest.Weight == "" {
		t.Error("PlumbusRequest.Weight should not be empty")
	}
	if validRequest.Wrapping == "" {
		t.Error("PlumbusRequest.Wrapping should not be empty")
	}
}

func TestPlumbusGenerationRequest_Fields(t *testing.T) {
	req := PlumbusGenerationRequest{
		Size:     "large",
		Color:    "red",
		Shape:    "square",
		Weight:   "heavy",
		Wrapping: "premium",
	}

	if req.Size != "large" {
		t.Errorf("Size = %v, want large", req.Size)
	}
	if req.Color != "red" {
		t.Errorf("Color = %v, want red", req.Color)
	}
	if req.Shape != "square" {
		t.Errorf("Shape = %v, want square", req.Shape)
	}
	if req.Weight != "heavy" {
		t.Errorf("Weight = %v, want heavy", req.Weight)
	}
	if req.Wrapping != "premium" {
		t.Errorf("Wrapping = %v, want premium", req.Wrapping)
	}
}
