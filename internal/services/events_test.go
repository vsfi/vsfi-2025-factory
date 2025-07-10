package services

import (
	"encoding/json"
	"testing"
	"time"

	"factory/internal/config"
	"factory/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MockNATSConn - мок для NATS соединения
type MockNATSConn struct {
	PublishedMessages []MockMessage
	ShouldFailPublish bool
	ShouldFailConnect bool
	IsClosed          bool
}

type MockMessage struct {
	Subject string
	Data    []byte
}

func (m *MockNATSConn) Publish(subject string, data []byte) error {
	if m.ShouldFailPublish {
		return &mockNATSError{msg: "failed to publish"}
	}

	m.PublishedMessages = append(m.PublishedMessages, MockMessage{
		Subject: subject,
		Data:    data,
	})
	return nil
}

// Убеждаемся, что MockNATSConn реализует интерфейс NATSConn
var _ NATSConn = (*MockNATSConn)(nil)

func (m *MockNATSConn) Close() {
	m.IsClosed = true
}

type mockNATSError struct {
	msg string
}

func (e *mockNATSError) Error() string {
	return e.msg
}

func TestEventsService_PublishPlumbusCreated_Success(t *testing.T) {
	cfg := &config.Config{
		NatsTopic:   "test-topic",
		EventSource: "test-factory",
	}

	// Создаем мок NATS соединения
	mockConn := &MockNATSConn{}

	service := &EventsService{
		conn:   mockConn,
		config: cfg,
		logger: logrus.New(),
	}

	// Создаем тестовые данные
	userID := uuid.New()
	plumbusID := uuid.New()

	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	plumbus := &models.Plumbus{
		ID:     plumbusID,
		UserID: userID,
		IsRare: true,
	}

	request := models.PlumbusRequest{
		Name:     "Test Plumbus",
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	// Публикуем событие
	err := service.PublishPlumbusCreated(user, plumbus, request)

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("PublishPlumbusCreated() error = %v, want nil", err)
	}

	// Проверяем что сообщение было опубликовано
	if len(mockConn.PublishedMessages) != 1 {
		t.Fatalf("PublishPlumbusCreated() published %d messages, want 1", len(mockConn.PublishedMessages))
	}

	message := mockConn.PublishedMessages[0]

	// Проверяем топик
	if message.Subject != "test-topic" {
		t.Errorf("PublishPlumbusCreated() subject = %v, want test-topic", message.Subject)
	}

	// Парсим событие
	var event PlumbusCreatedEvent
	err = json.Unmarshal(message.Data, &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal published event: %v", err)
	}

	// Проверяем поля события
	if event.Type != "plumbus.created" {
		t.Errorf("Event type = %v, want plumbus.created", event.Type)
	}

	if event.Source != "test-factory" {
		t.Errorf("Event source = %v, want test-factory", event.Source)
	}

	if event.ID == "" {
		t.Error("Event ID should not be empty")
	}

	if event.Timestamp.IsZero() {
		t.Error("Event timestamp should not be zero")
	}

	// Проверяем данные события
	data := event.Data
	if data["plumbus_id"] != plumbusID.String() {
		t.Errorf("Event data plumbus_id = %v, want %v", data["plumbus_id"], plumbusID)
	}

	if data["user_id"] != userID.String() {
		t.Errorf("Event data user_id = %v, want %v", data["user_id"], userID)
	}

	if data["username"] != "testuser" {
		t.Errorf("Event data username = %v, want testuser", data["username"])
	}

	if data["email"] != "test@example.com" {
		t.Errorf("Event data email = %v, want test@example.com", data["email"])
	}

	if data["is_rare"] != true {
		t.Errorf("Event data is_rare = %v, want true", data["is_rare"])
	}

	// Проверяем что plumbus_data содержит запрос
	plumbusData, ok := data["plumbus_data"].(map[string]interface{})
	if !ok {
		t.Fatal("Event data plumbus_data is not a map")
	}

	if plumbusData["name"] != "Test Plumbus" {
		t.Errorf("Event plumbus_data name = %v, want Test Plumbus", plumbusData["name"])
	}
}

func TestEventsService_PublishPlumbusCreated_PublishError(t *testing.T) {
	cfg := &config.Config{
		NatsTopic:   "test-topic",
		EventSource: "test-factory",
	}

	// Создаем мок NATS соединения с ошибкой публикации
	mockConn := &MockNATSConn{
		ShouldFailPublish: true,
	}

	service := &EventsService{
		conn:   mockConn,
		config: cfg,
		logger: logrus.New(),
	}

	// Создаем тестовые данные
	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	plumbus := &models.Plumbus{
		ID:     uuid.New(),
		UserID: user.ID,
		IsRare: false,
	}

	request := models.PlumbusRequest{
		Name:     "Test Plumbus",
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	// Публикуем событие
	err := service.PublishPlumbusCreated(user, plumbus, request)

	// Проверяем что вернулась ошибка
	if err == nil {
		t.Error("PublishPlumbusCreated() expected error for publish failure, got nil")
	}

	if err.Error() != "failed to publish event: failed to publish" {
		t.Errorf("PublishPlumbusCreated() error = %v, want failed to publish event: failed to publish", err)
	}
}

func TestEventsService_Close(t *testing.T) {
	// Создаем мок NATS соединения
	mockConn := &MockNATSConn{}

	service := &EventsService{
		conn:   mockConn,
		config: &config.Config{},
		logger: logrus.New(),
	}

	// Закрываем сервис
	service.Close()

	// Проверяем что соединение закрыто
	if !mockConn.IsClosed {
		t.Error("Close() did not close NATS connection")
	}
}

func TestEventsService_Close_NilConnection(t *testing.T) {
	service := &EventsService{
		conn:   nil,
		config: &config.Config{},
		logger: logrus.New(),
	}

	// Закрываем сервис с nil соединением (не должно паниковать)
	service.Close()
	// Тест проходит если нет паники
}

func TestPlumbusCreatedEvent_Structure(t *testing.T) {
	event := PlumbusCreatedEvent{
		ID:        "test-id",
		Type:      "plumbus.created",
		Source:    "factory",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plumbus_id": "test-plumbus-id",
			"user_id":    "test-user-id",
		},
	}

	// Проверяем что событие можно сериализовать в JSON
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal PlumbusCreatedEvent: %v", err)
	}

	// Проверяем что событие можно десериализовать из JSON
	var unmarshaled PlumbusCreatedEvent
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal PlumbusCreatedEvent: %v", err)
	}

	// Проверяем что значения корректны
	if unmarshaled.ID != event.ID {
		t.Errorf("Unmarshaled ID = %v, want %v", unmarshaled.ID, event.ID)
	}

	if unmarshaled.Type != event.Type {
		t.Errorf("Unmarshaled Type = %v, want %v", unmarshaled.Type, event.Type)
	}

	if unmarshaled.Source != event.Source {
		t.Errorf("Unmarshaled Source = %v, want %v", unmarshaled.Source, event.Source)
	}
}
