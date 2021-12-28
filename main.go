package main

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/wilbertthelam/prop-ock/handlers/health"
	"github.com/wilbertthelam/prop-ock/handlers/message"
	"github.com/wilbertthelam/prop-ock/handlers/webview"
)

func main() {
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

	if l, ok := e.Logger.(*log.Logger); ok {
		l.SetLevel(log.INFO)
	}

	root := InitializeDependencyInjectedModules()

	// Routes
	e.GET("/health", root.healthHandler.GetHealthCheck)

	e.POST("/message/send/auction", root.messageHandler.SendMessage)
	e.POST("/message/create/league", root.messageHandler.CreateLeague)
	e.POST("/message/create/auction", root.messageHandler.CreateAuction)
	e.GET("/message/webhook", root.messageHandler.VerifyMessengerWebhook)
	e.POST("/message/webhook", root.messageHandler.ProcessMessengerWebhook)

	// Templates
	e.File("/webview/bid", "public/bid.html")
	// e.GET("/webview/bid", root.webviewHandler.RenderBid)

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
}

type Root struct {
	healthHandler  *health.HealthHandler
	messageHandler *message.MessageHandler
	webviewHandler *webview.WebviewHandler
}

func New(
	healthHandler *health.HealthHandler,
	messageHandler *message.MessageHandler,
	webviewHandler *webview.WebviewHandler,
) *Root {
	return &Root{
		healthHandler,
		messageHandler,
		webviewHandler,
	}
}
