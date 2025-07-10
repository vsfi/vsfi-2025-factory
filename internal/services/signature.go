package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"factory/internal/config"
)

type SignatureService struct {
	config *config.Config
	client *http.Client
}

type SignatureResponse struct {
	CreatedAt    time.Time `json:"created_at"`
	SerialNumber int64     `json:"id"`
	Signature    string    `json:"signature"`
}

func NewSignatureService(cfg *config.Config) *SignatureService {
	return &SignatureService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *SignatureService) SignFile(filePath string) (*SignatureResponse, error) {
	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Создаем multipart форму
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Добавляем файл в форму
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file to form: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Отправляем запрос к sig-store
	url := fmt.Sprintf("%s/api/v1/register", s.config.SigStoreURL)
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sig-store returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Парсим ответ
	var sigResponse SignatureResponse
	err = json.NewDecoder(resp.Body).Decode(&sigResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &sigResponse, nil
}

func (s *SignatureService) VerifySignature(filePath, signature string) (bool, error) {
	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Создаем multipart форму
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Добавляем файл в форму
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return false, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return false, fmt.Errorf("failed to copy file to form: %w", err)
	}

	// Добавляем подпись в форму
	err = writer.WriteField("signature", signature)
	if err != nil {
		return false, fmt.Errorf("failed to write signature field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return false, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Отправляем запрос к sig-store
	url := fmt.Sprintf("%s/api/v1/verify", s.config.SigStoreURL)
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("sig-store returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Парсим ответ
	var verifyResponse struct {
		Valid   bool   `json:"valid"`
		Message string `json:"message"`
	}
	err = json.NewDecoder(resp.Body).Decode(&verifyResponse)
	if err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return verifyResponse.Valid, nil
}
