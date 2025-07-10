package main

import (
	"io"
	"os"
	"time"

	"factory/internal/config"
	"factory/internal/database"
	"factory/internal/handlers"
	"factory/internal/keycloak"
	"factory/internal/logger"
	"factory/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Инициализируем структурированное логирование
	log := logger.Init()

	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found")
	}

	// Инициализируем конфигурацию
	cfg := config.New()

	// Настраиваем Gin для JSON логирования
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Используем logrus writer для Gin
	gin.DefaultWriter = io.MultiWriter(os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(os.Stdout)

	// Инициализируем базу данных
	db, err := database.Initialize(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database")
	}

	// Инициализируем Keycloak клиент
	kcClient := keycloak.NewClient(cfg)

	// Инициализируем сервисы
	plumbusService := services.NewPlumbusService(cfg)
	userService := services.NewUserService(db)
	signatureService := services.NewSignatureService(cfg)

	// Инициализируем сервис событий NATS
	eventsService, err := services.NewEventsService(cfg)
	if err != nil {
		log.WithError(err).Warn("Failed to initialize NATS events service")
		eventsService = nil // Продолжаем работу без NATS
	}

	// Закрываем соединение с NATS при завершении
	if eventsService != nil {
		defer eventsService.Close()
	}

	// Настраиваем роутер
	router := gin.New()

	// Добавляем middleware для JSON логирования
	router.Use(ginJSONLogger(log))
	router.Use(gin.Recovery())

	// Статические файлы (CSS, JS, изображения)
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// Инициализируем обработчики
	h := handlers.NewHandler(plumbusService, userService, signatureService, eventsService, kcClient)

	// Маршруты
	router.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})
	router.GET("/", h.HomePage)
	router.GET("/auth/login", h.Login)
	router.GET("/auth/callback", h.AuthCallback)
	router.GET("/auth/logout", h.Logout)

	// Защищенные маршруты
	protected := router.Group("/")
	protected.Use(h.AuthMiddleware())
	{
		protected.GET("/dashboard", h.Dashboard)
		protected.POST("/plumbus/generate", h.GeneratePlumbus)
		protected.GET("/plumbus/status/:id", h.GetPlumbusStatus)
		protected.GET("/plumbus/image/:id", h.GetPlumbusImage)
		protected.GET("/plumbus/list", h.GetUserPlumbuses)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.WithField("port", port).Info("Starting server")
	if err := router.Run(":" + port); err != nil {
		log.WithError(err).Fatal("Failed to start server")
	}
}

// ginJSONLogger возвращает middleware для JSON логирования Gin запросов
func ginJSONLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Обрабатываем запрос
		c.Next()

		// Логируем запрос
		end := time.Now()
		latency := end.Sub(start)

		entry := logger.WithFields(logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"latency":    latency.String(),
			"time":       end.Format("2006-01-02T15:04:05.000Z07:00"),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("Request processed")
		}
	}
}
