package player_service

import (
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
	player_repo "github.com/wilbertthelam/prop-ock/repos/player"
)

type PlayerService struct {
	playerRepo *player_repo.PlayerRepo
}

func New(playerRepo *player_repo.PlayerRepo) *PlayerService {
	return &PlayerService{
		playerRepo,
	}
}

func (p *PlayerService) GetPlayerByPlayerId(context echo.Context, playerId string) (entities.Player, error) {
	// return p.playerRepo.GetPlayerByPlayerId(context, playerId)

	if playerId == "44-julio-rodriguez" {
		return entities.Player{
			Id:       playerId,
			Name:     "Julio Rodriguez",
			Image:    "https://a.espncdn.com/combiner/i?img=/i/headshots/mlb/players/full/41044.png",
			Team:     "SEA",
			Position: "OF",
		}, nil
	} else if playerId == "7-jarred-kelenic" {
		return entities.Player{
			Id:       playerId,
			Name:     "Jarred Kelenic",
			Image:    "https://a.espncdn.com/combiner/i?img=/i/headshots/mlb/players/full/41150.png",
			Team:     "SEA",
			Position: "OF",
		}, nil
	} else if playerId == "13-bobby-witt" {
		return entities.Player{
			Id:       playerId,
			Name:     "Bobby Witt",
			Image:    "https://a.espncdn.com/combiner/i?img=/i/headshots/mlb/players/full/42403.png",
			Team:     "KC",
			Position: "SS",
		}, nil
	}

	return entities.Player{}, nil
}
