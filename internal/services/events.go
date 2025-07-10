package services

import (
	"encoding/json"
	"fmt"
	"time"

	"factory/internal/config"
	"factory/internal/logger"
	"factory/internal/models"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

type EventsService struct {
	conn   *nats.Conn
	config *config.Config
	logger *logrus.Logger
}

// PlumbusCreatedEvent представляет событие создания плюмбуса
// в формате совместимом с events-audit
type PlumbusCreatedEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func NewEventsService(cfg *config.Config) (*EventsService, error) {
	logger := logger.Init()

	// Подключаемся к NATS
	conn, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	logger.WithField("nats_url", cfg.NatsURL).Info("Connected to NATS")

	return &EventsService{
		conn:   conn,
		config: cfg,
		logger: logger,
	}, nil
}

// Close закрывает соединение с NATS
func (s *EventsService) Close() {
	if s.conn != nil {
		s.conn.Close()
		s.logger.Info("NATS connection closed")
	}
}

// PublishPlumbusCreated отправляет событие о создании плюмбуса в NATS
func (s *EventsService) PublishPlumbusCreated(user *models.User, plumbus *models.Plumbus, request models.PlumbusRequest) error {
	event := PlumbusCreatedEvent{
		ID:        uuid.New().String(),
		Type:      "plumbus.created",
		Source:    s.config.EventSource,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plumbus_id":   plumbus.ID,
			"user_id":      user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"plumbus_data": request,
			"is_rare":      plumbus.IsRare,
		},
	}

	// Сериализуем событие в JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Отправляем событие в NATS
	err = s.conn.Publish(s.config.NatsTopic, eventData)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"source":     event.Source,
		"plumbus_id": plumbus.ID,
		"user_id":    user.ID,
		"username":   user.Username,
		"is_rare":    plumbus.IsRare,
		"topic":      s.config.NatsTopic,
	}).Info("Published plumbus.created event")

	return nil
}
