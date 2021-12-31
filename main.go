package main

import (
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/wilbertthelam/prop-ock/handlers/auction"
	"github.com/wilbertthelam/prop-ock/handlers/health"
	"github.com/wilbertthelam/prop-ock/handlers/league"
	"github.com/wilbertthelam/prop-ock/handlers/message"
	"github.com/wilbertthelam/prop-ock/handlers/player"
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
	e.GET("/api/health", root.healthHandler.GetHealthCheck)

	// Auction
	e.POST("/api/auction/create", root.auctionHandler.CreateAuction)
	e.POST("/api/auction/stop", root.auctionHandler.StopAuction)
	e.POST("/api/auction/bid/make", root.auctionHandler.MakeBid)
	e.POST("/api/auction/bid/cancel", root.auctionHandler.CancelBid)
	e.POST("/api/auction/process", root.auctionHandler.ProcessAuction)
	e.GET("/api/auction/current", root.auctionHandler.GetCurrentAuctionForLeague)
	e.GET("/api/auction/results", root.auctionHandler.GetAuctionResults)
	// e.GET("/auction", root.auctionHandler.GetAuction)

	// League
	e.POST("/api/league/create", root.leagueHandler.CreateLeague)

	// Players
	e.GET("/api/player", root.playerHandler.GetPlayer)

	// Messenger
	e.POST("/message/auction/players", root.messageHandler.SendPlayersForBidding)
	e.POST("/message/auction/results", root.messageHandler.SendWinningBids)
	e.GET("/message/webhook", root.messageHandler.VerifyMessengerWebhook)
	e.POST("/message/webhook", root.messageHandler.ProcessMessengerWebhook)

	// Webview
	e.Static("/webview/bid", "public/bid")

	// Start server
	e.Logger.Fatal(e.Start(":" + port))
}

type Root struct {
	healthHandler  *health.HealthHandler
	messageHandler *message.MessageHandler
	webviewHandler *webview.WebviewHandler
	auctionHandler *auction.AuctionHandler
	leagueHandler  *league.LeagueHandler
	playerHandler  *player.PlayerHandler
}

func New(
	healthHandler *health.HealthHandler,
	messageHandler *message.MessageHandler,
	webviewHandler *webview.WebviewHandler,
	auctionHandler *auction.AuctionHandler,
	leagueHandler *league.LeagueHandler,
	playerHandler *player.PlayerHandler,
) *Root {
	return &Root{
		healthHandler,
		messageHandler,
		webviewHandler,
		auctionHandler,
		leagueHandler,
		playerHandler,
	}
}
