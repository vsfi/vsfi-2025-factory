package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestUser_Fields(t *testing.T) {
	id := uuid.New()
	user := &User{
		ID:         id,
		KeycloakID: "test-keycloak-id",
		Username:   "testuser",
		Email:      "test@example.com",
	}

	// Проверяем что поля установлены корректно
	if user.ID != id {
		t.Errorf("User.ID = %v, want %v", user.ID, id)
	}

	if user.KeycloakID != "test-keycloak-id" {
		t.Errorf("User.KeycloakID = %v, want test-keycloak-id", user.KeycloakID)
	}

	if user.Username != "testuser" {
		t.Errorf("User.Username = %v, want testuser", user.Username)
	}

	if user.Email != "test@example.com" {
		t.Errorf("User.Email = %v, want test@example.com", user.Email)
	}
}

func TestPlumbus_Fields(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	plumbus := &Plumbus{
		ID:       id,
		UserID:   userID,
		Name:     "Test Plumbus",
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
		Status:   StatusPending,
	}

	// Проверяем что поля установлены корректно
	if plumbus.ID != id {
		t.Errorf("Plumbus.ID = %v, want %v", plumbus.ID, id)
	}

	if plumbus.UserID != userID {
		t.Errorf("Plumbus.UserID = %v, want %v", plumbus.UserID, userID)
	}

	if plumbus.Name != "Test Plumbus" {
		t.Errorf("Plumbus.Name = %v, want Test Plumbus", plumbus.Name)
	}

	if plumbus.Size != "medium" {
		t.Errorf("Plumbus.Size = %v, want medium", plumbus.Size)
	}

	if plumbus.Color != "blue" {
		t.Errorf("Plumbus.Color = %v, want blue", plumbus.Color)
	}

	if plumbus.Shape != "round" {
		t.Errorf("Plumbus.Shape = %v, want round", plumbus.Shape)
	}

	if plumbus.Weight != "light" {
		t.Errorf("Plumbus.Weight = %v, want light", plumbus.Weight)
	}

	if plumbus.Wrapping != "gift" {
		t.Errorf("Plumbus.Wrapping = %v, want gift", plumbus.Wrapping)
	}

	if plumbus.Status != StatusPending {
		t.Errorf("Plumbus.Status = %v, want %v", plumbus.Status, StatusPending)
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
