package config

import (
	"os"
	"testing"
)

func TestNew_DefaultValues(t *testing.T) {
	// Очищаем переменные окружения для тестирования значений по умолчанию
	envVars := []string{
		"DATABASE_URL",
		"KEYCLOAK_URL",
		"KEYCLOAK_INTERNAL_URL",
		"KEYCLOAK_REALM",
		"KEYCLOAK_CLIENT_ID",
		"KEYCLOAK_CLIENT_SECRET",
		"PLUMBUS_SERVICE_URL",
		"SIG_STORE_URL",
		"SESSION_SECRET",
		"NATS_URL",
		"NATS_TOPIC",
		"EVENT_SOURCE",
	}

	// Сохраняем текущие значения
	savedValues := make(map[string]string)
	for _, envVar := range envVars {
		savedValues[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}

	// Восстанавливаем значения после теста
	defer func() {
		for envVar, value := range savedValues {
			if value != "" {
				os.Setenv(envVar, value)
			}
		}
	}()

	cfg := New()

	// Проверяем значения по умолчанию
	tests := []struct {
		name     string
		actual   string
		expected string
	}{
		{"DatabaseURL", cfg.DatabaseURL, "postgres://postgres:accountant@localhost:5432/factory?sslmode=disable"},
		{"KeycloakURL", cfg.KeycloakURL, "http://localhost:8080"},
		{"KeycloakInternalURL", cfg.KeycloakInternalURL, "http://localhost:8080"},
		{"KeycloakRealm", cfg.KeycloakRealm, "master"},
		{"KeycloakClientID", cfg.KeycloakClientID, "factory"},
		{"KeycloakClientSecret", cfg.KeycloakClientSecret, ""},
		{"PlumbusServiceURL", cfg.PlumbusServiceURL, "http://localhost:8081"},
		{"SigStoreURL", cfg.SigStoreURL, "http://localhost:3000"},
		{"SessionSecret", cfg.SessionSecret, "your-secret-key"},
		{"NatsURL", cfg.NatsURL, "nats://localhost:4222"},
		{"NatsTopic", cfg.NatsTopic, "accountats"},
		{"EventSource", cfg.EventSource, "factory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.actual, tt.expected)
			}
		})
	}
}

func TestNew_EnvironmentValues(t *testing.T) {
	// Устанавливаем тестовые переменные окружения
	testValues := map[string]string{
		"DATABASE_URL":           "postgres://test:test@test:5432/test",
		"KEYCLOAK_URL":           "http://test:9090",
		"KEYCLOAK_INTERNAL_URL":  "http://test-internal:9090",
		"KEYCLOAK_REALM":         "test-realm",
		"KEYCLOAK_CLIENT_ID":     "test-client",
		"KEYCLOAK_CLIENT_SECRET": "test-secret",
		"PLUMBUS_SERVICE_URL":    "http://test:8081",
		"SIG_STORE_URL":          "http://test:3000",
		"SESSION_SECRET":         "test-session-secret",
		"NATS_URL":               "nats://test:4222",
		"NATS_TOPIC":             "test-topic",
		"EVENT_SOURCE":           "test-factory",
	}

	// Устанавливаем переменные окружения
	for key, value := range testValues {
		os.Setenv(key, value)
	}

	// Очищаем после теста
	defer func() {
		for key := range testValues {
			os.Unsetenv(key)
		}
	}()

	cfg := New()

	// Проверяем что значения загрузились из переменных окружения
	tests := []struct {
		name     string
		actual   string
		expected string
	}{
		{"DatabaseURL", cfg.DatabaseURL, testValues["DATABASE_URL"]},
		{"KeycloakURL", cfg.KeycloakURL, testValues["KEYCLOAK_URL"]},
		{"KeycloakInternalURL", cfg.KeycloakInternalURL, testValues["KEYCLOAK_INTERNAL_URL"]},
		{"KeycloakRealm", cfg.KeycloakRealm, testValues["KEYCLOAK_REALM"]},
		{"KeycloakClientID", cfg.KeycloakClientID, testValues["KEYCLOAK_CLIENT_ID"]},
		{"KeycloakClientSecret", cfg.KeycloakClientSecret, testValues["KEYCLOAK_CLIENT_SECRET"]},
		{"PlumbusServiceURL", cfg.PlumbusServiceURL, testValues["PLUMBUS_SERVICE_URL"]},
		{"SigStoreURL", cfg.SigStoreURL, testValues["SIG_STORE_URL"]},
		{"SessionSecret", cfg.SessionSecret, testValues["SESSION_SECRET"]},
		{"NatsURL", cfg.NatsURL, testValues["NATS_URL"]},
		{"NatsTopic", cfg.NatsTopic, testValues["NATS_TOPIC"]},
		{"EventSource", cfg.EventSource, testValues["EVENT_SOURCE"]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.actual, tt.expected)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	// Тест с существующей переменной окружения
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := getEnv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("getEnv() = %v, want test_value", result)
	}

	// Тест с несуществующей переменной окружения
	result = getEnv("NON_EXISTENT_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("getEnv() = %v, want default_value", result)
	}

	// Тест с пустой переменной окружения
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result = getEnv("EMPTY_VAR", "default")
	if result != "default" {
		t.Errorf("getEnv() with empty env var = %v, want default", result)
	}
}
