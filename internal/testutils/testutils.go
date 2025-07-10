package testutils

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// SQLiteUUID - пользовательский тип для работы с UUID в SQLite
type SQLiteUUID uuid.UUID

// Value реализует интерфейс driver.Valuer
func (u SQLiteUUID) Value() (driver.Value, error) {
	return uuid.UUID(u).String(), nil
}

// Scan реализует интерфейс sql.Scanner
func (u *SQLiteUUID) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	switch val := src.(type) {
	case string:
		parsed, err := uuid.Parse(val)
		if err != nil {
			return err
		}
		*u = SQLiteUUID(parsed)
		return nil
	case []byte:
		parsed, err := uuid.Parse(string(val))
		if err != nil {
			return err
		}
		*u = SQLiteUUID(parsed)
		return nil
	default:
		return fmt.Errorf("unsupported type for UUID: %T", src)
	}
}

// GormDataType реализует интерфейс schema.GormDataTypeInterface
func (SQLiteUUID) GormDataType() string {
	return "text"
}

// MockRoundTripper - мок для HTTP клиента
type MockRoundTripper struct {
	ResponseMap map[string]*http.Response
	RequestLog  []*http.Request
}

// NewMockRoundTripper создает новый мок для HTTP клиента
func NewMockRoundTripper() *MockRoundTripper {
	return &MockRoundTripper{
		ResponseMap: make(map[string]*http.Response),
		RequestLog:  make([]*http.Request, 0),
	}
}

// RoundTrip реализует интерфейс http.RoundTripper
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Сохраняем запрос для проверки в тестах
	m.RequestLog = append(m.RequestLog, req)

	// Ищем заранее подготовленный ответ
	key := req.Method + " " + req.URL.String()
	if resp, exists := m.ResponseMap[key]; exists {
		return resp, nil
	}

	// Возвращаем ошибку 404 если ответ не найден
	return &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader("Not Found")),
		Header:     make(http.Header),
	}, nil
}

// AddResponse добавляет мок-ответ для HTTP запроса
func (m *MockRoundTripper) AddResponse(method, url string, statusCode int, body string) {
	key := method + " " + url
	m.ResponseMap[key] = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// AddJSONResponse добавляет мок-ответ с JSON данными
func (m *MockRoundTripper) AddJSONResponse(method, url string, statusCode int, body string) {
	key := method + " " + url
	header := make(http.Header)
	header.Set("Content-Type", "application/json")

	m.ResponseMap[key] = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     header,
	}
}

// AddFileResponse добавляет мок-ответ с бинарными данными (например, изображение)
func (m *MockRoundTripper) AddFileResponse(method, url string, statusCode int, data []byte) {
	key := method + " " + url
	header := make(http.Header)
	header.Set("Content-Type", "image/png")

	m.ResponseMap[key] = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     header,
	}
}

// GetLastRequest возвращает последний выполненный запрос
func (m *MockRoundTripper) GetLastRequest() *http.Request {
	if len(m.RequestLog) == 0 {
		return nil
	}
	return m.RequestLog[len(m.RequestLog)-1]
}

// GetRequestCount возвращает количество выполненных запросов
func (m *MockRoundTripper) GetRequestCount() int {
	return len(m.RequestLog)
}

// Reset очищает лог запросов и ответов
func (m *MockRoundTripper) Reset() {
	m.RequestLog = make([]*http.Request, 0)
	m.ResponseMap = make(map[string]*http.Response)
}

// CreateTestPNGData создает тестовые данные PNG файла
func CreateTestPNGData() []byte {
	// Минимальный PNG файл (1x1 пиксель, прозрачный)
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR chunk size
		0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x02, 0x00, 0x00, 0x00, // bit depth, color type, compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
		0x00, 0x00, 0x00, 0x0C, // IDAT chunk size
		0x49, 0x44, 0x41, 0x54, // IDAT
		0x08, 0x99, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // IEND chunk size
		0x49, 0x45, 0x4E, 0x44, // IEND
		0xAE, 0x42, 0x60, 0x82, // CRC
	}
}

// SetupTestDB создает тестовую базу данных SQLite в памяти
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		DryRun: false,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Создаем схему для тестовой базы данных
	err = createTestSchema(db)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

// createTestSchema создает схему для тестовой базы данных
func createTestSchema(db *gorm.DB) error {
	// Создаем таблицу users
	err := db.Exec(`
		CREATE TABLE IF NOT EXISTS "user" (
			id TEXT PRIMARY KEY,
			keycloak_id TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		return err
	}

	// Создаем таблицу plumbuses
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS "plumbus" (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			size TEXT NOT NULL,
			color TEXT NOT NULL,
			shape TEXT NOT NULL,
			weight TEXT NOT NULL,
			wrapping TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			is_rare INTEGER NOT NULL DEFAULT 0,
			image_path TEXT,
			signature TEXT,
			signature_date DATETIME,
			error_msg TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES "user"(id)
		)
	`).Error
	if err != nil {
		return err
	}

	// Создаем триггер для обновления updated_at в users
	err = db.Exec(`
		CREATE TRIGGER IF NOT EXISTS users_updated_at
		AFTER UPDATE ON "user"
		BEGIN
			UPDATE "user" 
			SET updated_at = CURRENT_TIMESTAMP
			WHERE id = NEW.id;
		END;
	`).Error
	if err != nil {
		return err
	}

	// Создаем триггер для обновления updated_at в plumbuses
	err = db.Exec(`
		CREATE TRIGGER IF NOT EXISTS plumbuses_updated_at
		AFTER UPDATE ON "plumbus"
		BEGIN
			UPDATE "plumbus" 
			SET updated_at = CURRENT_TIMESTAMP
			WHERE id = NEW.id;
		END;
	`).Error
	if err != nil {
		return err
	}

	// Создаем триггер для генерации UUID при вставке в users
	err = db.Exec(`
		CREATE TRIGGER IF NOT EXISTS users_uuid_trigger
		AFTER INSERT ON "user"
		WHEN new.id IS NULL
		BEGIN
			UPDATE "user" 
			SET id = lower(
				hex(randomblob(4)) || '-' ||
				hex(randomblob(2)) || '-4' ||
				substr(hex(randomblob(2)), 2) || '-' ||
				substr('89ab', abs(random() % 4) + 1, 1) ||
				substr(hex(randomblob(2)), 2) || '-' ||
				hex(randomblob(6))
			)
			WHERE rowid = new.rowid;
		END;
	`).Error
	if err != nil {
		return err
	}

	// Создаем триггер для генерации UUID при вставке в plumbuses
	err = db.Exec(`
		CREATE TRIGGER IF NOT EXISTS plumbuses_uuid_trigger
		AFTER INSERT ON "plumbus"
		WHEN new.id IS NULL
		BEGIN
			UPDATE "plumbus" 
			SET id = lower(
				hex(randomblob(4)) || '-' ||
				hex(randomblob(2)) || '-4' ||
				substr(hex(randomblob(2)), 2) || '-' ||
				substr('89ab', abs(random() % 4) + 1, 1) ||
				substr(hex(randomblob(2)), 2) || '-' ||
				hex(randomblob(6))
			)
			WHERE rowid = new.rowid;
		END;
	`).Error
	if err != nil {
		return err
	}

	return nil
}
