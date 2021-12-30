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

func generateLeagueRedisKey(leagueId uuid.UUID) string {
	return fmt.Sprintf("league:league_id:%v", leagueId.String())
}

func generateLeagueMembersRelationshipKey(leagueId uuid.UUID) string {
	return fmt.Sprintf("relationship:league_to_user:%v", leagueId.String())
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

func (l *LeagueRepo) IsUserMemberOfLeague(context echo.Context, userId uuid.UUID, leagueId uuid.UUID) (bool, error) {
	isMember, err := l.redisClient.SIsMember(
		context.Request().Context(),
		generateLeagueMembersRelationshipKey(leagueId),
		userId.String(),
	).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if user is in league: userId: %v, leagueId: %v", userId, leagueId)
	}

	return isMember, nil
}

func (l *LeagueRepo) AddUserToLeague(context echo.Context, userId uuid.UUID, leagueId uuid.UUID) error {
	_, err := l.redisClient.SAdd(
		context.Request().Context(),
		generateLeagueMembersRelationshipKey(leagueId),
		userId.String(),
	).Result()
	if err != nil {
		return err
	}

	return nil
}

func (l *LeagueRepo) GetMembersInLeague(context echo.Context, leagueId uuid.UUID) ([]uuid.UUID, error) {
	stringUserIds, err := l.redisClient.SMembers(
		context.Request().Context(),
		generateLeagueMembersRelationshipKey(leagueId),
	).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get users in league: leagueId: %v", leagueId)
	}

	userIds := make([]uuid.UUID, len(stringUserIds))
	for index, stringUserId := range stringUserIds {
		userId, err := uuid.Parse(stringUserId)
		if err != nil {
			return nil, fmt.Errorf("failed to parse userId from Redis string to uuid: %v", stringUserId)
		}

		userIds[index] = userId
	}

	return userIds, nil
}
