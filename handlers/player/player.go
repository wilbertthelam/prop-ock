package player

import (
	"net/http"

	"github.com/labstack/echo/v4"
	player_service "github.com/wilbertthelam/prop-ock/services/player"
)

type PlayerHandler struct {
	playerService *player_service.PlayerService
}

func New(playerService *player_service.PlayerService) *PlayerHandler {
	return &PlayerHandler{
		playerService,
	}
}

func (p *PlayerHandler) GetPlayer(context echo.Context) error {
	player, err := p.playerService.GetPlayerByPlayerId(context, context.QueryParam("playerId"))
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}
	return context.JSON(http.StatusOK, player)
}
