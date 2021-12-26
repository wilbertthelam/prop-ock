package league_repo

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
)

type LeagueRepo struct {
	redisClient *redis.Client
}

func New(redisClient *redis.Client) *LeagueRepo {
	return &LeagueRepo{
		redisClient,
	}
}

func GetName() string {
	return "league_repo"
}

func generateLeagueRedisKey(leagueId uuid.UUID) string {
	return fmt.Sprintf("league:league_id:%v", leagueId.String())
}

func (l *LeagueRepo) GetLeagueByLeagueId(context echo.Context, leagueId uuid.UUID) (entities.League, error) {
	redisLeague, err := l.redisClient.HGetAll(
		context.Request().Context(),
		generateLeagueRedisKey(leagueId),
	).Result()
	if err != nil {
		return entities.League{}, fmt.Errorf("failed to get league: %v", leagueId)
	}

	// If league is not found, then return an empty league
	if len(redisLeague) == 0 {
		return entities.League{}, nil
	}

	league := entities.League{
		Id:   uuid.Must(uuid.Parse(redisLeague["id"])),
		Name: redisLeague["name"],
	}

	return league, nil
}

func (l *LeagueRepo) CreateLeague(context echo.Context, leagueId uuid.UUID, league entities.League) error {
	// Format for database inserting (array of strings where even index is key, odd index is value)
	redisLeagueKeyValuePairs := []string{
		"id", league.Id.String(),
		"name", league.Name,
	}

	err := l.updateLeague(context, leagueId, redisLeagueKeyValuePairs)
	if err != nil {
		return err
	}

	return nil
}

func (l *LeagueRepo) updateLeague(context echo.Context, leagueId uuid.UUID, keyValuePairs []string) error {
	_, err := l.redisClient.HSet(
		context.Request().Context(),
		generateLeagueRedisKey(leagueId),
		keyValuePairs,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to update league fields: error: %+v, leagueId: %v, keyValuePairs: %+v", err, leagueId, keyValuePairs)
	}

	return nil
}
