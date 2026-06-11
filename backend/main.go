package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"hrms-backend/internal/config"
	"hrms-backend/internal/database"
	"hrms-backend/internal/handlers"
	"hrms-backend/internal/middleware"
	"hrms-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL()); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize services
	authService := service.NewAuthService(cfg)
	employeeService := service.NewEmployeeService()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	employeeHandler := handlers.NewEmployeeHandler(employeeService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:       "HRMS API",
		CaseSensitive: true,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(helmet.New())
	app.Use(middleware.CORSConfig(cfg))

	// Health check
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"data": fiber.Map{
				"status":  "healthy",
				"version": "1.0.0",
			},
		})
	})

	// Public routes (no auth required)
	api := app.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)

	// Protected routes (auth required)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(authService))

	// User info
	protected.Get("/auth/me", authHandler.Me)

	// Employee routes
	employees := protected.Group("/employees")
	employees.Get("/", employeeHandler.ListEmployees)
	employees.Get("/:id", employeeHandler.GetEmployee)

	// Dashboard
	protected.Get("/dashboard", employeeHandler.Dashboard)

	// Start server
	go func() {
		addr := cfg.ServerHost + ":" + cfg.ServerPort
		log.Printf("Server starting on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
}
