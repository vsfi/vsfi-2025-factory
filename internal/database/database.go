package database

import (
	"factory/internal/config"
	"factory/internal/models"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Initialize(cfg *config.Config) (*gorm.DB, error) {
	// Сначала создаем БД если не существует
	err := EnsureDatabaseExists(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure database exists: %w", err)
	}

	// Настраиваем уровень логирования GORM
	var logLevel logger.LogLevel
	envLogLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))

	switch envLogLevel {
	case "trace", "debug":
		logLevel = logger.Info // Показываем SQL запросы в debug/trace режиме
	case "warn", "warning":
		logLevel = logger.Warn
	case "error":
		logLevel = logger.Error
	default:
		logLevel = logger.Silent // По умолчанию скрываем SQL запросы
	}

	// Теперь подключаемся к целевой БД
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	// Автомиграция
	err = db.AutoMigrate(
		&models.User{},
		&models.Plumbus{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func EnsureDatabaseExists(databaseURL string) error {
	// Парсим URL для получения параметров подключения
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Извлекаем имя БД из пути
	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" {
		return fmt.Errorf("database name not specified in URL")
	}

	// Создаем URL для подключения к postgres БД (системной)
	systemDBURL := strings.Replace(databaseURL, "/"+dbName, "/postgres", 1)

	// Подключаемся к системной БД
	systemDB, err := gorm.Open(postgres.Open(systemDBURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Всегда скрываем логи системной БД
	})
	if err != nil {
		return fmt.Errorf("failed to connect to system database: %w", err)
	}

	// Получаем нативное соединение для выполнения SQL запросов
	sqlDB, err := systemDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB: %w", err)
	}
	defer sqlDB.Close()

	// Проверяем существование БД
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = systemDB.Raw(query, dbName).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	// Создаем БД если не существует
	if !exists {
		log.Printf("Database '%s' does not exist, creating...", dbName)
		createQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
		err = systemDB.Exec(createQuery).Error
		if err != nil {
			return fmt.Errorf("failed to create database '%s': %w", dbName, err)
		}
		log.Printf("Database '%s' created successfully", dbName)
	} else {
		log.Printf("Database '%s' already exists", dbName)
	}

	return nil
}
