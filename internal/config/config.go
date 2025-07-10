package config

import "os"

type Config struct {
	DatabaseURL          string
	KeycloakURL          string
	KeycloakInternalURL  string
	KeycloakRealm        string
	KeycloakClientID     string
	KeycloakClientSecret string
	PlumbusServiceURL    string
	SigStoreURL          string
	SessionSecret        string
	NatsURL              string
	NatsTopic            string
	EventSource          string
}

func New() *Config {
	return &Config{
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:accountant@localhost:5432/factory?sslmode=disable"),
		KeycloakURL:          getEnv("KEYCLOAK_URL", "http://localhost:8080"),
		KeycloakInternalURL:  getEnv("KEYCLOAK_INTERNAL_URL", "http://localhost:8080"),
		KeycloakRealm:        getEnv("KEYCLOAK_REALM", "master"),
		KeycloakClientID:     getEnv("KEYCLOAK_CLIENT_ID", "factory"),
		KeycloakClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", ""),
		PlumbusServiceURL:    getEnv("PLUMBUS_SERVICE_URL", "http://localhost:8081"),
		SigStoreURL:          getEnv("SIG_STORE_URL", "http://localhost:3000"),
		SessionSecret:        getEnv("SESSION_SECRET", "your-secret-key"),
		NatsURL:              getEnv("NATS_URL", "nats://localhost:4222"),
		NatsTopic:            getEnv("NATS_TOPIC", "accountats"),
		EventSource:          getEnv("EVENT_SOURCE", "factory"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
