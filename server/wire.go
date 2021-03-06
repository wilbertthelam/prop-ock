//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	redis_client "github.com/wilbertthelam/prop-ock/db"
	"github.com/wilbertthelam/prop-ock/handlers/auction"
	"github.com/wilbertthelam/prop-ock/handlers/health"
	"github.com/wilbertthelam/prop-ock/handlers/league"
	"github.com/wilbertthelam/prop-ock/handlers/message"
	"github.com/wilbertthelam/prop-ock/handlers/player"
	"github.com/wilbertthelam/prop-ock/handlers/webview"
	auction_repo "github.com/wilbertthelam/prop-ock/repos/auction"
	league_repo "github.com/wilbertthelam/prop-ock/repos/league"
	player_repo "github.com/wilbertthelam/prop-ock/repos/player"
	user_repo "github.com/wilbertthelam/prop-ock/repos/user"
	auction_service "github.com/wilbertthelam/prop-ock/services/auction"
	callups_service "github.com/wilbertthelam/prop-ock/services/callups"
	config_service "github.com/wilbertthelam/prop-ock/services/config"
	league_service "github.com/wilbertthelam/prop-ock/services/league"
	message_service "github.com/wilbertthelam/prop-ock/services/message"
	player_service "github.com/wilbertthelam/prop-ock/services/player"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
)

func InitializeDependencyInjectedModules() *Root {
	wire.Build(
		New,
		health.New,
		webview.New,
		message.New,
		player.New,
		league.New,
		auction.New,
		auction_service.New,
		callups_service.New,
		user_service.New,
		league_service.New,
		message_service.New,
		player_service.New,
		auction_repo.New,
		league_repo.New,
		player_repo.New,
		user_repo.New,
		redis_client.New,
		config_service.New,
	)

	return &Root{}
}
