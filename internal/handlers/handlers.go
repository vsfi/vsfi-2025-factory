package handlers

import (
	"context"
	"fmt"
	"net/http"

	"factory/internal/keycloak"
	"factory/internal/logger"
	"factory/internal/models"
	"factory/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	plumbusService   *services.PlumbusService
	userService      *services.UserService
	signatureService *services.SignatureService
	eventsService    *services.EventsService
	keycloakClient   *keycloak.Client
	logger           *logrus.Logger
}

func NewHandler(ps *services.PlumbusService, us *services.UserService, ss *services.SignatureService, es *services.EventsService, kc *keycloak.Client) *Handler {
	return &Handler{
		plumbusService:   ps,
		userService:      us,
		signatureService: ss,
		eventsService:    es,
		keycloakClient:   kc,
		logger:           logger.Init(),
	}
}

func (h *Handler) HomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Rick & Morty Plumbus Factory",
	})
}

func (h *Handler) Login(c *gin.Context) {
	redirectURI := fmt.Sprintf("http://%s/auth/callback", c.Request.Host)
	loginURL := h.keycloakClient.GetLoginURL(redirectURI)
	c.Redirect(http.StatusTemporaryRedirect, loginURL)
}

func (h *Handler) AuthCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		h.logger.WithField("error", "missing_authorization_code").Error("Auth callback failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	redirectURI := fmt.Sprintf("http://%s/auth/callback", c.Request.Host)

	token, err := h.keycloakClient.ExchangeCodeForToken(context.Background(), code, redirectURI)
	if err != nil {
		h.logger.WithError(err).Error("Failed to exchange code for token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	// Получаем информацию о пользователе
	tokenPreview := token.AccessToken
	if len(tokenPreview) > 50 {
		tokenPreview = tokenPreview[:50] + "..."
	}
	h.logger.WithField("token_preview", tokenPreview).Debug("Token received")

	userInfo, err := h.keycloakClient.VerifyToken(context.Background(), token.AccessToken)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user info")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	h.logger.WithField("user_info", userInfo).Debug("UserInfo received")

	// Проверяем, что все необходимые поля присутствуют
	if userInfo.Sub == nil || userInfo.PreferredUsername == nil || userInfo.Email == nil {
		h.logger.WithFields(logrus.Fields{
			"sub":      userInfo.Sub,
			"username": userInfo.PreferredUsername,
			"email":    userInfo.Email,
		}).Error("Missing required user info fields")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Incomplete user information"})
		return
	}

	// Создаем или получаем пользователя в БД
	user, err := h.userService.GetOrCreateUser(*userInfo.Sub, *userInfo.PreferredUsername, *userInfo.Email)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create/get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Сохраняем токен в сессии
	c.SetCookie("access_token", token.AccessToken, 3600, "/", "", false, true)
	c.SetCookie("user_id", user.ID.String(), 3600, "/", "", false, false)

	h.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
	}).Info("User authenticated successfully")

	c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
}

func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("user_id", "", -1, "/", "", false, false)
	h.logger.Info("User logged out")
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}

		// Верифицируем токен
		_, err = h.keycloakClient.VerifyToken(context.Background(), token)
		if err != nil {
			h.logger.WithError(err).Debug("Token verification failed")
			c.SetCookie("access_token", "", -1, "/", "", false, true)
			c.SetCookie("user_id", "", -1, "/", "", false, false)
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *Handler) Dashboard(c *gin.Context) {
	userIDStr, _ := c.Cookie("user_id")
	userID, _ := uuid.Parse(userIDStr)

	// Получаем информацию о пользователе
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user info")
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	plumbuses, err := h.userService.GetUserPlumbuses(userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user plumbuses")
		plumbuses = []models.Plumbus{}
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":     "Dashboard - Rick & Morty Plumbus Factory",
		"user":      user,
		"plumbuses": plumbuses,
	})
}

func (h *Handler) GeneratePlumbus(c *gin.Context) {
	var req models.PlumbusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Invalid plumbus request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Cookie("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).Error("Invalid user ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Создаем запись плюмбуса в БД
	plumbus, err := h.userService.CreatePlumbus(userID, req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"request": req,
		}).Error("Failed to create plumbus")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"plumbus_id": plumbus.ID,
		"user_id":    userID,
		"is_rare":    plumbus.IsRare,
		"name":       req.Name,
	}).Info("Plumbus created successfully")

	// Отправляем событие о создании плюмбуса в NATS
	if h.eventsService != nil {
		// Получаем информацию о пользователе для события
		user, err := h.userService.GetUserByID(userID)
		if err != nil {
			h.logger.WithError(err).WithField("user_id", userID).Warn("Failed to get user info for event")
		} else {
			// Отправляем событие в горутине чтобы не блокировать основной поток
			go func() {
				if err := h.eventsService.PublishPlumbusCreated(user, plumbus, req); err != nil {
					h.logger.WithError(err).WithField("plumbus_id", plumbus.ID).Warn("Failed to publish plumbus created event")
				}
			}()
		}
	}

	// Запускаем генерацию в горутине
	go h.generatePlumbusAsync(plumbus.ID, models.PlumbusGenerationRequest{
		Size:     req.Size,
		Color:    req.Color,
		Shape:    req.Shape,
		Weight:   req.Weight,
		Wrapping: req.Wrapping,
	})

	c.JSON(http.StatusOK, gin.H{
		"id":      plumbus.ID,
		"status":  "generating",
		"is_rare": plumbus.IsRare,
	})
}

func (h *Handler) generatePlumbusAsync(plumbusID uuid.UUID, req models.PlumbusGenerationRequest) {
	// Обновляем статус на "generating"
	h.userService.UpdatePlumbusStatus(plumbusID, models.StatusGenerating, nil, nil, nil, nil)

	h.logger.WithFields(logrus.Fields{
		"plumbus_id": plumbusID,
		"request":    req,
	}).Info("Starting plumbus generation")

	// Генерируем плюмбус
	imagePath, err := h.plumbusService.GeneratePlumbus(req)
	if err != nil {
		h.logger.WithError(err).WithField("plumbus_id", plumbusID).Error("Failed to generate plumbus")
		errorMsg := err.Error()
		h.userService.UpdatePlumbusStatus(plumbusID, models.StatusFailed, nil, &errorMsg, nil, nil)
		return
	}

	// Подписываем изображение плюмбуса
	h.logger.WithFields(logrus.Fields{
		"plumbus_id": plumbusID,
		"image_path": imagePath,
	}).Info("Signing plumbus image")

	signatureResponse, err := h.signatureService.SignFile(imagePath)
	if err != nil {
		h.logger.WithError(err).WithField("plumbus_id", plumbusID).Error("Failed to sign plumbus image")
		// Не считаем это критической ошибкой, продолжаем без подписи
		h.userService.UpdatePlumbusStatus(plumbusID, models.StatusCompleted, &imagePath, nil, nil, nil)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"plumbus_id":    plumbusID,
		"signature":     signatureResponse.Signature[:20] + "...",
		"serial_number": signatureResponse.SerialNumber,
	}).Info("Plumbus signed successfully")

	// Обновляем статус на "completed" с подписью
	h.userService.UpdatePlumbusStatus(plumbusID, models.StatusCompleted, &imagePath, nil,
		&signatureResponse.Signature, &signatureResponse.CreatedAt)
}

func (h *Handler) GetPlumbusStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).WithField("id_string", idStr).Error("Invalid plumbus ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	plumbus, err := h.userService.GetPlumbus(id)
	if err != nil {
		h.logger.WithError(err).WithField("plumbus_id", id).Error("Plumbus not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Plumbus not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             plumbus.ID,
		"status":         plumbus.Status,
		"name":           plumbus.Name,
		"is_rare":        plumbus.IsRare,
		"signature":      plumbus.Signature,
		"signature_date": plumbus.SignatureDate,
	})
}

func (h *Handler) GetPlumbusImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.WithError(err).WithField("id_string", idStr).Error("Invalid plumbus ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	plumbus, err := h.userService.GetPlumbus(id)
	if err != nil {
		h.logger.WithError(err).WithField("plumbus_id", id).Error("Plumbus not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Plumbus not found"})
		return
	}

	if plumbus.Status != models.StatusCompleted || plumbus.ImagePath == nil {
		h.logger.WithFields(logrus.Fields{
			"plumbus_id": id,
			"status":     plumbus.Status,
			"has_image":  plumbus.ImagePath != nil,
		}).Warn("Image not available")
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not available"})
		return
	}

	c.File(*plumbus.ImagePath)
}

func (h *Handler) GetUserPlumbuses(c *gin.Context) {
	userIDStr, _ := c.Cookie("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.WithError(err).WithField("user_id_string", userIDStr).Error("Invalid user ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	plumbuses, err := h.userService.GetUserPlumbuses(userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user plumbuses")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, plumbuses)
}
