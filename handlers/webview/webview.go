package webview

import (
	auction_service "github.com/wilbertthelam/prop-ock/services/auction"
	player_service "github.com/wilbertthelam/prop-ock/services/player"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
)

type WebviewHandler struct {
	playerService  *player_service.PlayerService
	auctionService *auction_service.AuctionService
	userService    *user_service.UserService
}

func New(
	playerService *player_service.PlayerService,
	auctionService *auction_service.AuctionService,
	userService *user_service.UserService,
) *WebviewHandler {
	return &WebviewHandler{
		playerService,
		auctionService,
		userService,
	}
}

func GetName() string {
	return "webview"
}
