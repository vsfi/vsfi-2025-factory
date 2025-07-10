package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"factory/internal/config"
	"factory/internal/models"

	"github.com/google/uuid"
)

type PlumbusService struct {
	config *config.Config
	client *http.Client
}

func NewPlumbusService(cfg *config.Config) *PlumbusService {
	return &PlumbusService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *PlumbusService) GeneratePlumbus(req models.PlumbusGenerationRequest) (string, error) {
	// Подготавливаем запрос к сервису генерации
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Отправляем запрос
	url := fmt.Sprintf("%s/plumbus", s.config.PlumbusServiceURL)
	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("service returned status %d", resp.StatusCode)
	}

	// Создаем папку для хранения изображений если её нет
	imgDir := "storage/images"
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create image directory: %w", err)
	}

	// Генерируем уникальное имя файла
	fileName := fmt.Sprintf("%s.png", uuid.New().String())
	filePath := filepath.Join(imgDir, fileName)

	// Сохраняем изображение
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return filePath, nil
}
