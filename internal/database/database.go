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
	"gorm.io/gorm/schema"
)

func Initialize(cfg *config.Config) (*gorm.DB, error) {
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

	// Парсим URL для получения параметров подключения
	parsedURL, err := url.Parse(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Извлекаем имя БД из пути
	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" {
		return nil, fmt.Errorf("database name not specified in URL")
	}

	// Создаем URL для подключения к postgres БД (системной)
	systemDBURL := strings.Replace(cfg.DatabaseURL, "/"+dbName, "/postgres", 1)

	// Подключаемся к системной БД
	systemDB, err := gorm.Open(postgres.Open(systemDBURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Всегда скрываем логи системной БД
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system database: %w", err)
	}

	// Создаем расширение uuid-ossp в системной БД
	if err := systemDB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" WITH SCHEMA public;").Error; err != nil {
		return nil, fmt.Errorf("failed to create uuid extension in system database: %w", err)
	}

	// Проверяем существование БД
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = systemDB.Raw(query, dbName).Scan(&exists).Error
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	// Создаем БД если не существует
	if !exists {
		log.Printf("Database '%s' does not exist, creating...", dbName)
		createQuery := fmt.Sprintf("CREATE DATABASE %s WITH TEMPLATE template0", dbName)
		err = systemDB.Exec(createQuery).Error
		if err != nil {
			return nil, fmt.Errorf("failed to create database '%s': %w", dbName, err)
		}
		log.Printf("Database '%s' created successfully", dbName)
	} else {
		log.Printf("Database '%s' already exists", dbName)
	}

	// Закрываем соединение с системной БД
	sqlDB, err := systemDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get system DB connection: %w", err)
	}
	sqlDB.Close()

	// Теперь подключаемся к целевой БД с дополнительными настройками
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // Используем имена таблиц в единственном числе
		},
	})
	if err != nil {
		return nil, err
	}

	// Создаем расширение uuid-ossp в целевой БД
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\" WITH SCHEMA public;").Error; err != nil {
		return nil, fmt.Errorf("failed to create uuid extension in target database: %w", err)
	}

	// Проверяем, что расширение создано
	var hasExtension bool
	err = db.Raw("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'uuid-ossp')").Scan(&hasExtension).Error
	if err != nil {
		return nil, fmt.Errorf("failed to check uuid extension: %w", err)
	}
	if !hasExtension {
		return nil, fmt.Errorf("uuid-ossp extension was not created")
	}

	log.Printf("Starting database migration...")

	// Проверяем существование таблиц и создаем их, если они отсутствуют
	if !db.Migrator().HasTable(&models.User{}) {
		log.Printf("Creating users table...")
		if err := db.Migrator().CreateTable(&models.User{}); err != nil {
			return nil, fmt.Errorf("failed to create users table: %w", err)
		}
	} else {
		log.Printf("Users table already exists")
	}

	if !db.Migrator().HasTable(&models.Plumbus{}) {
		log.Printf("Creating plumbuses table...")
		if err := db.Migrator().CreateTable(&models.Plumbus{}); err != nil {
			return nil, fmt.Errorf("failed to create plumbuses table: %w", err)
		}
	} else {
		log.Printf("Plumbuses table already exists")
	}

	// Проверяем, что таблицы существуют
	log.Printf("Verifying table existence...")
	if !db.Migrator().HasTable(&models.User{}) {
		return nil, fmt.Errorf("users table was not created")
	}
	if !db.Migrator().HasTable(&models.Plumbus{}) {
		return nil, fmt.Errorf("plumbuses table was not created")
	}

	log.Printf("Database migration completed successfully")

	return db, nil
}
