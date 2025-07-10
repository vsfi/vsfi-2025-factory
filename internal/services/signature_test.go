package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"factory/internal/config"
	"factory/internal/testutils"
)

func TestNewSignatureService(t *testing.T) {
	cfg := &config.Config{
		SigStoreURL: "http://localhost:3000",
	}

	service := NewSignatureService(cfg)

	if service == nil {
		t.Fatal("NewSignatureService() returned nil")
	}

	if service.config != cfg {
		t.Error("NewSignatureService() did not set config correctly")
	}

	if service.client == nil {
		t.Error("NewSignatureService() did not initialize HTTP client")
	}

	if service.client.Timeout != 30*time.Second {
		t.Errorf("NewSignatureService() timeout = %v, want %v", service.client.Timeout, 30*time.Second)
	}
}

func TestSignatureService_SignFile_Success(t *testing.T) {
	// Создаем временный файл для тестирования
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.png")
	testContent := testutils.CreateTestPNGData()

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Настраиваем мок HTTP клиента
	mockRT := testutils.NewMockRoundTripper()
	mockResponse := `{
		"created_at": "2023-12-01T10:00:00Z",
		"id": 12345,
		"signature": "test-signature-hash"
	}`
	mockRT.AddJSONResponse("POST", "http://localhost:3000/api/v1/register", 200, mockResponse)

	cfg := &config.Config{
		SigStoreURL: "http://localhost:3000",
	}

	service := NewSignatureService(cfg)
	service.client.Transport = mockRT

	// Выполняем подписание файла
	result, err := service.SignFile(testFile)

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("SignFile() error = %v, want nil", err)
	}

	// Проверяем результат
	if result == nil {
		t.Fatal("SignFile() returned nil result")
	}

	if result.SerialNumber != 12345 {
		t.Errorf("SignFile() SerialNumber = %v, want 12345", result.SerialNumber)
	}

	if result.Signature != "test-signature-hash" {
		t.Errorf("SignFile() Signature = %v, want test-signature-hash", result.Signature)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2023-12-01T10:00:00Z")
	if !result.CreatedAt.Equal(expectedTime) {
		t.Errorf("SignFile() CreatedAt = %v, want %v", result.CreatedAt, expectedTime)
	}

	// Проверяем что запрос был отправлен правильно
	lastReq := mockRT.GetLastRequest()
	if lastReq == nil {
		t.Fatal("No HTTP request was made")
	}

	if lastReq.Method != "POST" {
		t.Errorf("Request method = %v, want POST", lastReq.Method)
	}

	if lastReq.URL.String() != "http://localhost:3000/api/v1/register" {
		t.Errorf("Request URL = %v, want http://localhost:3000/api/v1/register", lastReq.URL.String())
	}

	if !strings.Contains(lastReq.Header.Get("Content-Type"), "multipart/form-data") {
		t.Errorf("Request Content-Type = %v, want multipart/form-data", lastReq.Header.Get("Content-Type"))
	}
}

func TestSignatureService_SignFile_FileNotFound(t *testing.T) {
	cfg := &config.Config{
		SigStoreURL: "http://localhost:3000",
	}

	service := NewSignatureService(cfg)

	// Пытаемся подписать несуществующий файл
	_, err := service.SignFile("/nonexistent/file.png")

	// Проверяем что вернулась ошибка
	if err == nil {
		t.Error("SignFile() expected error for nonexistent file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("SignFile() error = %v, want error containing 'failed to open file'", err)
	}
}

func TestSignatureService_SignFile_ServiceError(t *testing.T) {
	// Создаем временный файл для тестирования
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.png")
	testContent := testutils.CreateTestPNGData()

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Настраиваем мок HTTP клиента с ошибкой
	mockRT := testutils.NewMockRoundTripper()
	mockRT.AddResponse("POST", "http://localhost:3000/api/v1/register", 500, "Internal Server Error")

	cfg := &config.Config{
		SigStoreURL: "http://localhost:3000",
	}

	service := NewSignatureService(cfg)
	service.client.Transport = mockRT

	// Выполняем подписание файла
	_, err = service.SignFile(testFile)

	// Проверяем что вернулась ошибка
	if err == nil {
		t.Error("SignFile() expected error for 500 status, got nil")
	}

	if !strings.Contains(err.Error(), "sig-store returned status 500") {
		t.Errorf("SignFile() error = %v, want error containing 'sig-store returned status 500'", err)
	}
}

func TestSignatureService_VerifySignature_Success(t *testing.T) {
	// Создаем временный файл для тестирования
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.png")
	testContent := testutils.CreateTestPNGData()

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Настраиваем мок HTTP клиента
	mockRT := testutils.NewMockRoundTripper()
	mockResponse := `{
		"valid": true,
		"message": "Signature is valid"
	}`
	mockRT.AddJSONResponse("POST", "http://localhost:3000/api/v1/verify", 200, mockResponse)

	cfg := &config.Config{
		SigStoreURL: "http://localhost:3000",
	}

	service := NewSignatureService(cfg)
	service.client.Transport = mockRT

	// Выполняем верификацию подписи
	isValid, err := service.VerifySignature(testFile, "test-signature")

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("VerifySignature() error = %v, want nil", err)
	}

	// Проверяем результат
	if !isValid {
		t.Error("VerifySignature() = false, want true")
	}

	// Проверяем что запрос был отправлен правильно
	lastReq := mockRT.GetLastRequest()
	if lastReq == nil {
		t.Fatal("No HTTP request was made")
	}

	if lastReq.Method != "POST" {
		t.Errorf("Request method = %v, want POST", lastReq.Method)
	}

	if lastReq.URL.String() != "http://localhost:3000/api/v1/verify" {
		t.Errorf("Request URL = %v, want http://localhost:3000/api/v1/verify", lastReq.URL.String())
	}

	if !strings.Contains(lastReq.Header.Get("Content-Type"), "multipart/form-data") {
		t.Errorf("Request Content-Type = %v, want multipart/form-data", lastReq.Header.Get("Content-Type"))
	}
}

func TestSignatureService_VerifySignature_Invalid(t *testing.T) {
	// Создаем временный файл для тестирования
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.png")
	testContent := testutils.CreateTestPNGData()

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Настраиваем мок HTTP клиента с невалидной подписью
	mockRT := testutils.NewMockRoundTripper()
	mockResponse := `{
		"valid": false,
		"message": "Signature is invalid"
	}`
	mockRT.AddJSONResponse("POST", "http://localhost:3000/api/v1/verify", 200, mockResponse)

	cfg := &config.Config{
		SigStoreURL: "http://localhost:3000",
	}

	service := NewSignatureService(cfg)
	service.client.Transport = mockRT

	// Выполняем верификацию подписи
	isValid, err := service.VerifySignature(testFile, "invalid-signature")

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("VerifySignature() error = %v, want nil", err)
	}

	// Проверяем результат
	if isValid {
		t.Error("VerifySignature() = true, want false")
	}
}

func TestSignatureService_VerifySignature_FileNotFound(t *testing.T) {
	cfg := &config.Config{
		SigStoreURL: "http://localhost:3000",
	}

	service := NewSignatureService(cfg)

	// Пытаемся верифицировать несуществующий файл
	_, err := service.VerifySignature("/nonexistent/file.png", "test-signature")

	// Проверяем что вернулась ошибка
	if err == nil {
		t.Error("VerifySignature() expected error for nonexistent file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("VerifySignature() error = %v, want error containing 'failed to open file'", err)
	}
}

func TestSignatureService_VerifySignature_ServiceError(t *testing.T) {
	// Создаем временный файл для тестирования
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.png")
	testContent := testutils.CreateTestPNGData()

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Настраиваем мок HTTP клиента с ошибкой
	mockRT := testutils.NewMockRoundTripper()
	mockRT.AddResponse("POST", "http://localhost:3000/api/v1/verify", 500, "Internal Server Error")

	cfg := &config.Config{
		SigStoreURL: "http://localhost:3000",
	}

	service := NewSignatureService(cfg)
	service.client.Transport = mockRT

	// Выполняем верификацию подписи
	_, err = service.VerifySignature(testFile, "test-signature")

	// Проверяем что вернулась ошибка
	if err == nil {
		t.Error("VerifySignature() expected error for 500 status, got nil")
	}

	if !strings.Contains(err.Error(), "sig-store returned status 500") {
		t.Errorf("VerifySignature() error = %v, want error containing 'sig-store returned status 500'", err)
	}
}
