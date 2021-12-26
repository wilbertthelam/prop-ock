package main

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/wilbertthelam/prop-ock/handlers/health"
	"github.com/wilbertthelam/prop-ock/handlers/message"
)

func main() {
	modules := LoadModules()

	// Get Port
	port := os.Getenv("PORT")

	// If no port (local dev), default to 8000
	if port == "" {
		port = "8000"
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	healthHandlerModule := modules[health.GetName()].(*health.HealthHandler)
	messageHandlerModule := modules[message.GetName()].(*message.MessageHandler)

	// Routes
	e.GET("/health", healthHandlerModule.GetHealthCheck)
	e.GET("/message/send", messageHandlerModule.SendMessage)
	e.GET("/message/get-latest", messageHandlerModule.GetLatestMessage)
	e.GET("/message/webhook", messageHandlerModule.VerifyMessengerWebhook)
	e.POST("/message/webhook", messageHandlerModule.ProcessMessengerWebhook)

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
}
