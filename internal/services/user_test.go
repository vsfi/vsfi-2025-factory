package services

import (
	"testing"
	"time"

	"factory/internal/models"
	"factory/internal/testutils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	return testutils.SetupTestDB(t)
}

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	id := uuid.New()
	if db.Name() == "sqlite" {
		sqliteUser := SQLiteUser{
			ID:         testutils.SQLiteUUID(id),
			KeycloakID: "test-keycloak-id",
			Username:   "testuser",
			Email:      "test@example.com",
		}
		if err := db.Create(&sqliteUser).Error; err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
		return &models.User{
			ID:         id,
			KeycloakID: sqliteUser.KeycloakID,
			Username:   sqliteUser.Username,
			Email:      sqliteUser.Email,
			CreatedAt:  sqliteUser.CreatedAt,
			UpdatedAt:  sqliteUser.UpdatedAt,
		}
	}

	user := models.User{
		ID:         id,
		KeycloakID: "test-keycloak-id",
		Username:   "testuser",
		Email:      "test@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return &user
}

func createTestPlumbus(t *testing.T, db *gorm.DB, userID uuid.UUID) *models.Plumbus {
	id := uuid.New()
	if db.Name() == "sqlite" {
		sqlitePlumbus := SQLitePlumbus{
			ID:       testutils.SQLiteUUID(id),
			UserID:   testutils.SQLiteUUID(userID),
			Name:     "Test Plumbus",
			Size:     "medium",
			Color:    "blue",
			Shape:    "round",
			Weight:   "light",
			Wrapping: "gift",
			Status:   models.StatusPending,
		}
		if err := db.Create(&sqlitePlumbus).Error; err != nil {
			t.Fatalf("Failed to create test plumbus: %v", err)
		}
		return &models.Plumbus{
			ID:        id,
			UserID:    userID,
			Name:      sqlitePlumbus.Name,
			Size:      sqlitePlumbus.Size,
			Color:     sqlitePlumbus.Color,
			Shape:     sqlitePlumbus.Shape,
			Weight:    sqlitePlumbus.Weight,
			Wrapping:  sqlitePlumbus.Wrapping,
			Status:    sqlitePlumbus.Status,
			CreatedAt: sqlitePlumbus.CreatedAt,
			UpdatedAt: sqlitePlumbus.UpdatedAt,
		}
	}

	plumbus := models.Plumbus{
		ID:       id,
		UserID:   userID,
		Name:     "Test Plumbus",
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
		Status:   models.StatusPending,
	}
	if err := db.Create(&plumbus).Error; err != nil {
		t.Fatalf("Failed to create test plumbus: %v", err)
	}
	return &plumbus
}

func TestNewUserService(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	if service == nil {
		t.Fatal("NewUserService() returned nil")
	}

	if service.db != db {
		t.Error("NewUserService() did not set database correctly")
	}
}

func TestUserService_GetOrCreateUser_CreateNew(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	keycloakID := "test-keycloak-id"
	username := "testuser"
	email := "test@example.com"

	// Создаем нового пользователя
	user, err := service.GetOrCreateUser(keycloakID, username, email)

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v, want nil", err)
	}

	// Проверяем что пользователь создался
	if user == nil {
		t.Fatal("GetOrCreateUser() returned nil user")
	}

	if user.KeycloakID != keycloakID {
		t.Errorf("User KeycloakID = %v, want %v", user.KeycloakID, keycloakID)
	}

	if user.Username != username {
		t.Errorf("User Username = %v, want %v", user.Username, username)
	}

	if user.Email != email {
		t.Errorf("User Email = %v, want %v", user.Email, email)
	}

	if user.ID == uuid.Nil {
		t.Error("User ID should not be nil")
	}
}

func TestUserService_GetOrCreateUser_GetExisting(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	// Создаем пользователя напрямую в базе
	existingUser := createTestUser(t, db)

	// Пытаемся получить/создать пользователя
	user, err := service.GetOrCreateUser(existingUser.KeycloakID, "newusername", "newemail@example.com")

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("GetOrCreateUser() error = %v, want nil", err)
	}

	// Проверяем что вернулся существующий пользователь
	if user.ID != existingUser.ID {
		t.Errorf("GetOrCreateUser() returned new user, want existing user")
	}

	if user.Username != existingUser.Username {
		t.Errorf("User Username = %v, want %v (original)", user.Username, existingUser.Username)
	}

	if user.Email != existingUser.Email {
		t.Errorf("User Email = %v, want %v (original)", user.Email, existingUser.Email)
	}
}

func TestUserService_CreatePlumbus(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	// Создаем тестового пользователя
	user := createTestUser(t, db)

	req := models.PlumbusRequest{
		Name:     "Test Plumbus",
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	// Создаем плюмбус
	plumbus, err := service.CreatePlumbus(user.ID, req)

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("CreatePlumbus() error = %v, want nil", err)
	}

	// Проверяем что плюмбус создался
	if plumbus == nil {
		t.Fatal("CreatePlumbus() returned nil plumbus")
	}

	if plumbus.ID == uuid.Nil {
		t.Error("Plumbus ID should not be nil")
	}

	if plumbus.UserID != user.ID {
		t.Errorf("Plumbus UserID = %v, want %v", plumbus.UserID, user.ID)
	}

	if plumbus.Name != req.Name {
		t.Errorf("Plumbus Name = %v, want %v", plumbus.Name, req.Name)
	}

	if plumbus.Size != req.Size {
		t.Errorf("Plumbus Size = %v, want %v", plumbus.Size, req.Size)
	}

	if plumbus.Color != req.Color {
		t.Errorf("Plumbus Color = %v, want %v", plumbus.Color, req.Color)
	}

	if plumbus.Shape != req.Shape {
		t.Errorf("Plumbus Shape = %v, want %v", plumbus.Shape, req.Shape)
	}

	if plumbus.Weight != req.Weight {
		t.Errorf("Plumbus Weight = %v, want %v", plumbus.Weight, req.Weight)
	}

	if plumbus.Wrapping != req.Wrapping {
		t.Errorf("Plumbus Wrapping = %v, want %v", plumbus.Wrapping, req.Wrapping)
	}

	if plumbus.Status != models.StatusPending {
		t.Errorf("Plumbus Status = %v, want %v", plumbus.Status, models.StatusPending)
	}
}

func TestUserService_UpdatePlumbusStatus(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	// Создаем тестового пользователя и плюмбус
	user := createTestUser(t, db)
	plumbus := createTestPlumbus(t, db, user.ID)

	imagePath := "test/path.png"
	signature := "test-signature"
	signatureDate := time.Now()

	// Обновляем статус плюмбуса
	err := service.UpdatePlumbusStatus(plumbus.ID, models.StatusCompleted, &imagePath, nil, &signature, &signatureDate)

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("UpdatePlumbusStatus() error = %v, want nil", err)
	}

	// Проверяем что плюмбус обновился в базе
	updatedPlumbus, err := service.GetPlumbus(plumbus.ID)
	if err != nil {
		t.Fatalf("Failed to fetch updated plumbus: %v", err)
	}

	if updatedPlumbus.Status != models.StatusCompleted {
		t.Errorf("Updated Plumbus Status = %v, want %v", updatedPlumbus.Status, models.StatusCompleted)
	}

	if updatedPlumbus.ImagePath == nil || *updatedPlumbus.ImagePath != imagePath {
		t.Errorf("Updated Plumbus ImagePath = %v, want %v", updatedPlumbus.ImagePath, imagePath)
	}

	if updatedPlumbus.Signature == nil || *updatedPlumbus.Signature != signature {
		t.Errorf("Updated Plumbus Signature = %v, want %v", updatedPlumbus.Signature, signature)
	}

	if updatedPlumbus.SignatureDate == nil || !updatedPlumbus.SignatureDate.Equal(signatureDate) {
		t.Errorf("Updated Plumbus SignatureDate = %v, want %v", updatedPlumbus.SignatureDate, signatureDate)
	}
}

func TestUserService_GetPlumbus(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	// Создаем тестового пользователя и плюмбус
	user := createTestUser(t, db)
	plumbus := createTestPlumbus(t, db, user.ID)

	// Получаем плюмбус
	retrievedPlumbus, err := service.GetPlumbus(plumbus.ID)

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("GetPlumbus() error = %v, want nil", err)
	}

	// Проверяем что плюмбус корректный
	if retrievedPlumbus.ID != plumbus.ID {
		t.Errorf("GetPlumbus() ID = %v, want %v", retrievedPlumbus.ID, plumbus.ID)
	}

	if retrievedPlumbus.Name != plumbus.Name {
		t.Errorf("GetPlumbus() Name = %v, want %v", retrievedPlumbus.Name, plumbus.Name)
	}
}

func TestUserService_GetPlumbus_NotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	nonExistentID := uuid.New()

	// Пытаемся получить несуществующий плюмбус
	_, err := service.GetPlumbus(nonExistentID)

	// Проверяем что вернулась ошибка
	if err == nil {
		t.Error("GetPlumbus() expected error for nonexistent plumbus, got nil")
	}

	// Проверяем что это именно ошибка "record not found"
	if err != gorm.ErrRecordNotFound {
		t.Errorf("GetPlumbus() error = %v, want %v", err, gorm.ErrRecordNotFound)
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	// Создаем тестового пользователя
	user := createTestUser(t, db)

	// Получаем пользователя
	retrievedUser, err := service.GetUserByID(user.ID)

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("GetUserByID() error = %v, want nil", err)
	}

	// Проверяем что пользователь корректный
	if retrievedUser.ID != user.ID {
		t.Errorf("GetUserByID() ID = %v, want %v", retrievedUser.ID, user.ID)
	}

	if retrievedUser.Username != user.Username {
		t.Errorf("GetUserByID() Username = %v, want %v", retrievedUser.Username, user.Username)
	}
}

func TestUserService_GetUserPlumbuses(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	// Создаем тестового пользователя
	user := createTestUser(t, db)

	// Создаем несколько плюмбусов для пользователя
	_ = createTestPlumbus(t, db, user.ID)
	_ = createTestPlumbus(t, db, user.ID)

	// Получаем плюмбусы пользователя
	plumbuses, err := service.GetUserPlumbuses(user.ID)

	// Проверяем что ошибки нет
	if err != nil {
		t.Fatalf("GetUserPlumbuses() error = %v, want nil", err)
	}

	// Проверяем что количество плюмбусов корректное
	if len(plumbuses) != 2 {
		t.Fatalf("GetUserPlumbuses() returned %d plumbuses, want 2", len(plumbuses))
	}

	// Проверяем что плюмбусы отсортированы по дате создания (DESC)
	if plumbuses[0].CreatedAt.Before(plumbuses[1].CreatedAt) {
		t.Error("GetUserPlumbuses() plumbuses are not sorted by created_at DESC")
	}

	// Проверяем что все плюмбусы принадлежат пользователю
	for _, plumbus := range plumbuses {
		if plumbus.UserID != user.ID {
			t.Errorf("Plumbus %s does not belong to user %s", plumbus.ID, user.ID)
		}
	}
}

func TestUserService_CreatePlumbus_RarenessProbability(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)

	// Создаем тестового пользователя
	user := createTestUser(t, db)

	req := models.PlumbusRequest{
		Name:     "Test Plumbus",
		Size:     "medium",
		Color:    "blue",
		Shape:    "round",
		Weight:   "light",
		Wrapping: "gift",
	}

	// Создаем много плюмбусов и проверяем что некоторые редкие
	rareCount := 0
	totalCount := 100

	for i := 0; i < totalCount; i++ {
		plumbus, err := service.CreatePlumbus(user.ID, req)
		if err != nil {
			t.Fatalf("CreatePlumbus() error = %v", err)
		}

		if plumbus.IsRare {
			rareCount++
		}
	}

	// Проверяем что редких плюмбусов не слишком много и не слишком мало
	// При вероятности 5% ожидаем примерно 5 редких из 100
	// Допускаем диапазон от 0 до 15 для статистической вариации
	if rareCount > 15 {
		t.Errorf("Too many rare plumbuses: %d out of %d (expected around 5)", rareCount, totalCount)
	}

	t.Logf("Created %d rare plumbuses out of %d total (%.1f%%)", rareCount, totalCount, float64(rareCount)/float64(totalCount)*100)
}
