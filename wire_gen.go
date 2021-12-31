// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/wilbertthelam/prop-ock/db"
	"github.com/wilbertthelam/prop-ock/handlers/auction"
	"github.com/wilbertthelam/prop-ock/handlers/health"
	"github.com/wilbertthelam/prop-ock/handlers/league"
	"github.com/wilbertthelam/prop-ock/handlers/message"
	"github.com/wilbertthelam/prop-ock/handlers/player"
	"github.com/wilbertthelam/prop-ock/handlers/webview"
	"github.com/wilbertthelam/prop-ock/repos/auction"
	"github.com/wilbertthelam/prop-ock/repos/league"
	"github.com/wilbertthelam/prop-ock/repos/player"
	"github.com/wilbertthelam/prop-ock/repos/user"
	"github.com/wilbertthelam/prop-ock/services/auction"
	"github.com/wilbertthelam/prop-ock/services/callups"
	"github.com/wilbertthelam/prop-ock/services/league"
	"github.com/wilbertthelam/prop-ock/services/message"
	"github.com/wilbertthelam/prop-ock/services/player"
	"github.com/wilbertthelam/prop-ock/services/user"
)

// Injectors from wire.go:

func InitializeDependencyInjectedModules() *Root {
	healthHandler := health.New()
	client := redis_client.New()
	auctionRepo := auction_repo.New(client)
	userRepo := user_repo.New(client)
	userService := user_service.New(userRepo)
	playerRepo := player_repo.New(client)
	playerService := player_service.New(playerRepo)
	leagueRepo := league_repo.New(client)
	leagueService := league_service.New(leagueRepo)
	auctionService := auction_service.New(auctionRepo, userService, playerService, leagueService)
	callupsService := callups_service.New(client)
	messageService := message_service.New(auctionService, userService, playerService, leagueService)
	messageHandler := message.New(auctionService, callupsService, userService, leagueService, messageService)
	webviewHandler := webview.New(playerService, auctionService, userService)
	auctionHandler := auction.New(auctionService, userService)
	leagueHandler := league.New(leagueService)
	playerHandler := player.New(playerService)
	root := New(healthHandler, messageHandler, webviewHandler, auctionHandler, leagueHandler, playerHandler)
	return root
}
