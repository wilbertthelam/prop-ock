package player_repo

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
	"github.com/wilbertthelam/prop-ock/utils"
)

type PlayerRepo struct {
	redisClient *redis.Client
}

func New(redisClient *redis.Client) *PlayerRepo {
	return &PlayerRepo{
		redisClient,
	}
}

func generatePlayerRedisKey(playerId string) string {
	return fmt.Sprintf("player:player_id:%v", playerId)
}

func (l *PlayerRepo) GetPlayerByPlayerId(context echo.Context, playerId string) (entities.Player, error) {
	redisPlayer, err := l.redisClient.HGetAll(
		context.Request().Context(),
		generatePlayerRedisKey(playerId),
	).Result()
	if err != nil {
		return entities.Player{}, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to get player",
			Args: []interface{}{
				"playerId", playerId,
			},
			Err: err,
		})
	}

	// If player is not found, then return an empty player
	if len(redisPlayer) == 0 {
		return entities.Player{}, nil
	}

	player := entities.Player{
		Id:    redisPlayer["id"],
		Name:  redisPlayer["name"],
		Image: redisPlayer["image"],
	}

	return player, nil
}

func (l *PlayerRepo) CreatePlayer(context echo.Context, playerId string, player entities.Player) error {
	// Format for database inserting (array of strings where even index is key, odd index is value)
	redisPlayerKeyValuePairs := []string{
		"id", player.Id,
		"name", player.Name,
		"image", player.Image,
	}

	err := l.updatePlayer(context, playerId, redisPlayerKeyValuePairs)
	if err != nil {
		return err
	}

	return nil
}

func (l *PlayerRepo) updatePlayer(context echo.Context, playerId string, keyValuePairs []string) error {
	_, err := l.redisClient.HSet(
		context.Request().Context(),
		generatePlayerRedisKey(playerId),
		keyValuePairs,
	).Result()
	if err != nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to update player fields",
			Args: append(
				[]interface{}{"playerId", playerId},
				utils.MapStringSliceToInterfaceSlice(keyValuePairs)...,
			),
			Err: err,
		})
	}

	return nil
}
