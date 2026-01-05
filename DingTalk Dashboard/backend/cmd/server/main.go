package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"dingtalk-dashboard/internal/config"
	"dingtalk-dashboard/internal/database"
	"dingtalk-dashboard/internal/dingtalk"
	"dingtalk-dashboard/internal/domain/approval"
	"dingtalk-dashboard/internal/handler"
	"dingtalk-dashboard/internal/middleware"
	"dingtalk-dashboard/internal/scheduler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("Failed to load config", zap.Error(err))
	}

	// Connect to database
	db, err := database.Connect(cfg, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize DingTalk client
	dtClient := dingtalk.NewClient(cfg.DingTalkAppKey, cfg.DingTalkAppSecret)

	// Initialize services
	approvalRepo := approval.NewRepository(db)
	approvalService := approval.NewService(approvalRepo, dtClient, zapLogger)

	// Initialize scheduler
	syncScheduler := scheduler.NewScheduler(
		approvalService,
		cfg.ApprovalProcessCode,
		cfg.Location,
		zapLogger,
	)

	// Start scheduler
	if err := syncScheduler.Start(); err != nil {
		zapLogger.Fatal("Failed to start scheduler", zap.Error(err))
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "DingTalk Dashboard API",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.NewCORS())

	// Health endpoints
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// API v1 routes
	v1 := app.Group("/api/v1")

	// Initialize handlers
	approvalHandler := handler.NewApprovalHandler(approvalService, syncScheduler)
	authHandler := handler.NewAuthHandler(cfg.AuthAPIBaseURL)

	// Determine JWT secret (prefer JWT_ACCESS_SECRET, fallback to JWT_SECRET)
	jwtSecret := cfg.JWTAccessSecret
	if jwtSecret == "" {
		jwtSecret = cfg.JWTSecret
	}

	// Auth middleware (optional - can be enabled/disabled)
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)

	// Auth proxy routes (public - handles CORS for external auth API)
	auth := v1.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/register", authHandler.Register)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", authHandler.Logout)

	// Approval routes (protected)
	approvals := v1.Group("/approvals")
	if jwtSecret != "" {
		approvals.Use(authMiddleware.Authenticate())
	}
	approvals.Get("/", approvalHandler.ListApprovals)
	approvals.Get("/stats", approvalHandler.GetStats)
	approvals.Get("/filter-options", approvalHandler.GetFilterOptions)
	approvals.Get("/:id", approvalHandler.GetApproval)

	// Sync routes (protected)
	sync := v1.Group("/sync")
	if jwtSecret != "" {
		sync.Use(authMiddleware.Authenticate())
	}
	sync.Get("/logs", approvalHandler.ListSyncLogs)
	sync.Post("/trigger", approvalHandler.TriggerSync)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		zapLogger.Info("Shutting down...")
		syncScheduler.Stop()
		app.ShutdownWithContext(context.Background())
	}()

	// Start server
	zapLogger.Info("Starting server", zap.String("port", cfg.Port))
	if err := app.Listen(":" + cfg.Port); err != nil {
		zapLogger.Fatal("Server failed", zap.Error(err))
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": err.Error(),
	})
}
