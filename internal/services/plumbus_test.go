package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"factory/internal/config"
	"factory/internal/models"
	"factory/internal/testutils"
)

func TestNewPlumbusService(t *testing.T) {
	cfg := &config.Config{
		PlumbusServiceURL: "http://localhost:8081",
	}

	service := NewPlumbusService(cfg)

	if service == nil {
		t.Fatal("NewPlumbusService() returned nil")
	}

	if service.config != cfg {
		t.Error("NewPlumbusService() did not set config correctly")
	}

	if service.client == nil {
		t.Error("NewPlumbusService() did not initialize HTTP client")
	}

	if service.client.Timeout != 30*time.Second {
		t.Errorf("NewPlumbusService() timeout = %v, want %v", service.client.Timeout, 30*time.Second)
	}
}

func TestPlumbusService_GeneratePlumbus_Success(t *testing.T) {
	// Создаем директорию для тестов
	testDir := t.TempDir()

	// Меняем рабочую директорию на тестовую
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Настраиваем мок HTTP клиента
	mockRT := testutils.NewMockRoundTripper()
	testPNG := testutils.CreateTestPNGData()
	mockRT.AddFileResponse("POST", "http://localhost:8081/plumbus", 200, testPNG)

	cfg := &config.Config{
		PlumbusServiceURL: "http://localhost:8081",
	}

	service := NewPlumbusService(cfg)
	service.client.Transport = mockRT

	req := models.PlumbusGenerationRequest{
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	filePath, err := service.GeneratePlumbus(req)

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("GeneratePlumbus() error = %v, want nil", err)
	}

	// Проверяем что путь к файлу корректный
	if filePath == "" {
		t.Error("GeneratePlumbus() returned empty file path")
	}

	// Проверяем что файл существует
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("GeneratePlumbus() file does not exist: %s", filePath)
	}

	// Проверяем что директория создалась
	expectedDir := "storage/images"
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("GeneratePlumbus() did not create directory: %s", expectedDir)
	}

	// Проверяем что запрос был отправлен правильно
	lastReq := mockRT.GetLastRequest()
	if lastReq == nil {
		t.Fatal("No HTTP request was made")
	}

	if lastReq.Method != "POST" {
		t.Errorf("Request method = %v, want POST", lastReq.Method)
	}

	if lastReq.URL.String() != "http://localhost:8081/plumbus" {
		t.Errorf("Request URL = %v, want http://localhost:8081/plumbus", lastReq.URL.String())
	}

	if lastReq.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Request Content-Type = %v, want application/json", lastReq.Header.Get("Content-Type"))
	}
}

func TestPlumbusService_GeneratePlumbus_ServiceError(t *testing.T) {
	// Настраиваем мок HTTP клиента с ошибкой
	mockRT := testutils.NewMockRoundTripper()
	mockRT.AddResponse("POST", "http://localhost:8081/plumbus", 500, "Internal Server Error")

	cfg := &config.Config{
		PlumbusServiceURL: "http://localhost:8081",
	}

	service := NewPlumbusService(cfg)
	service.client.Transport = mockRT

	req := models.PlumbusGenerationRequest{
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	_, err := service.GeneratePlumbus(req)

	// Проверяем что вернулась ошибка
	if err == nil {
		t.Error("GeneratePlumbus() expected error for 500 status, got nil")
	}

	// Проверяем что ошибка содержит статус код
	expectedError := "service returned status 500"
	if err.Error() != expectedError {
		t.Errorf("GeneratePlumbus() error = %v, want %v", err.Error(), expectedError)
	}
}

func TestPlumbusService_GeneratePlumbus_NetworkError(t *testing.T) {
	cfg := &config.Config{
		PlumbusServiceURL: "http://invalid-url-that-does-not-exist:9999",
	}

	service := NewPlumbusService(cfg)

	req := models.PlumbusGenerationRequest{
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	_, err := service.GeneratePlumbus(req)

	// Проверяем что вернулась ошибка сети
	if err == nil {
		t.Error("GeneratePlumbus() expected network error, got nil")
	}
}

func TestPlumbusService_GeneratePlumbus_FileCreationError(t *testing.T) {
	// Создаем временную директорию
	testDir := t.TempDir()

	// Меняем рабочую директорию на тестовую
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Создаем файл с именем storage чтобы нельзя было создать директорию
	err := os.WriteFile("storage", []byte("test"), 0644)
	if err != nil {
		t.Fatal("Failed to create test file")
	}

	// Настраиваем мок HTTP клиента
	mockRT := testutils.NewMockRoundTripper()
	testPNG := testutils.CreateTestPNGData()
	mockRT.AddFileResponse("POST", "http://localhost:8081/plumbus", 200, testPNG)

	cfg := &config.Config{
		PlumbusServiceURL: "http://localhost:8081",
	}

	service := NewPlumbusService(cfg)
	service.client.Transport = mockRT

	req := models.PlumbusGenerationRequest{
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	_, err = service.GeneratePlumbus(req)

	// Проверяем что вернулась ошибка создания директории
	if err == nil {
		t.Error("GeneratePlumbus() expected directory creation error, got nil")
	}
}

func TestPlumbusService_GeneratePlumbus_FilePathGeneration(t *testing.T) {
	// Создаем директорию для тестов
	testDir := t.TempDir()

	// Меняем рабочую директорию на тестовую
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(testDir)

	// Настраиваем мок HTTP клиента
	mockRT := testutils.NewMockRoundTripper()
	testPNG := testutils.CreateTestPNGData()
	mockRT.AddFileResponse("POST", "http://localhost:8081/plumbus", 200, testPNG)

	cfg := &config.Config{
		PlumbusServiceURL: "http://localhost:8081",
	}

	service := NewPlumbusService(cfg)
	service.client.Transport = mockRT

	req := models.PlumbusGenerationRequest{
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	// Генерируем несколько файлов
	paths := make([]string, 3)
	for i := 0; i < 3; i++ {
		path, err := service.GeneratePlumbus(req)
		if err != nil {
			t.Fatalf("GeneratePlumbus() error = %v", err)
		}
		paths[i] = path
	}

	// Проверяем что все пути разные (UUID должны быть уникальными)
	for i := 0; i < len(paths); i++ {
		for j := i + 1; j < len(paths); j++ {
			if paths[i] == paths[j] {
				t.Errorf("GeneratePlumbus() generated duplicate paths: %s", paths[i])
			}
		}
	}

	// Проверяем что все пути имеют правильный формат
	for _, path := range paths {
		dir := filepath.Dir(path)
		if dir != "storage/images" {
			t.Errorf("Generated path dir = %v, want storage/images", dir)
		}

		ext := filepath.Ext(path)
		if ext != ".png" {
			t.Errorf("Generated path ext = %v, want .png", ext)
		}
	}
}
