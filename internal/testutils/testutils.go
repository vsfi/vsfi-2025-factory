package testutils

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

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
