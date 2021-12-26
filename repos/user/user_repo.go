package user_repo

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
)

type UserRepo struct {
	redisClient *redis.Client
}

func New(redisClient *redis.Client) *UserRepo {
	return &UserRepo{
		redisClient,
	}
}

func GetName() string {
	return "user_repo"
}

func generateUserRedisKey(userId uuid.UUID) string {
	return fmt.Sprintf("user:user_id:%v", userId.String())
}

func generateUserWalletRedisKey(userId uuid.UUID) string {
	return fmt.Sprintf("wallet:user_id:%v", userId.String())
}

func generateLeagueMembersRelationshipKey(leagueId uuid.UUID) string {
	return fmt.Sprintf("relationship:league_to_user:%v", leagueId.String())
}

func generateSenderPsIdToUserIdRedisKey(senderPsId string) string {
	return fmt.Sprintf("relationship:sender_ps_id_to_user_id:%v", senderPsId)
}

func (u *UserRepo) GetUserByUserId(context echo.Context, userId uuid.UUID) (entities.User, error) {
	redisUser, err := u.redisClient.HGetAll(
		context.Request().Context(),
		generateUserRedisKey(userId),
	).Result()
	if err != nil {
		return entities.User{}, fmt.Errorf("failed to get user: %v", userId)
	}

	// If user is not found, then return an empty user
	if len(redisUser) == 0 {
		return entities.User{}, nil
	}

	user := entities.User{
		Id:   uuid.Must(uuid.Parse(redisUser["id"])),
		Name: redisUser["name"],
	}

	return user, nil
}

func (u *UserRepo) CreateUser(context echo.Context, userId uuid.UUID, user entities.User) error {
	// Format for database inserting (array of strings where even index is key, odd index is value)
	redisUserKeyValuePairs := []string{
		"id", user.Id.String(),
		"name", user.Name,
	}

	err := u.updateUser(context, userId, redisUserKeyValuePairs)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserRepo) IsUserMemberOfLeague(context echo.Context, userId uuid.UUID, leagueId uuid.UUID) (bool, error) {
	isMember, err := u.redisClient.SIsMember(
		context.Request().Context(),
		generateLeagueMembersRelationshipKey(leagueId),
		userId.String(),
	).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if user is in league: userId: %v, leagueId: %v", userId, leagueId)
	}

	return isMember, nil
}

func (u *UserRepo) AddUserToLeague(context echo.Context, userId uuid.UUID, leagueId uuid.UUID) error {
	_, err := u.redisClient.SAdd(
		context.Request().Context(),
		generateLeagueMembersRelationshipKey(leagueId),
		userId.String(),
	).Result()
	if err != nil {
		return err
	}

	return nil
}

func (u *UserRepo) updateUser(context echo.Context, userId uuid.UUID, keyValuePairs []string) error {
	_, err := u.redisClient.HSet(
		context.Request().Context(),
		generateUserRedisKey(userId),
		keyValuePairs,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to update user fields: error: %+v, userId: %v, keyValuePairs: %+v", err, userId, keyValuePairs)
	}

	return nil
}

func (u *UserRepo) GetUserWallet(context echo.Context, userId uuid.UUID) (map[uuid.UUID]int64, error) {
	wallet := make(map[uuid.UUID]int64)

	redisWallet, err := u.redisClient.HGetAll(
		context.Request().Context(),
		generateUserWalletRedisKey(userId),
	).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: error: %+v", err)
	}

	for key, value := range redisWallet {
		uuidKey, err := uuid.Parse(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user wallet leagueId key: %v", key)
		}

		int64Value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user wallet leagueId value: %v", value)
		}

		wallet[uuidKey] = int64Value
	}

	return wallet, nil
}

func (u *UserRepo) AddFundsToUserWallet(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (int64, error) {
	updatedWalletValue, err := u.incrementWalletFund(context, userId, leagueId, value)
	if err != nil {
		return 0, fmt.Errorf("failed to add funds to user wallet: error: %+v, fund value: %v, userId: %v, leagueId: %v", err, value, userId, leagueId)
	}

	return updatedWalletValue, nil
}

func (u *UserRepo) RemoveFundsFromUserWallet(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (int64, error) {
	updatedWalletValue, err := u.incrementWalletFund(context, userId, leagueId, value*-1)
	if err != nil {
		return 0, fmt.Errorf("failed to remove funds to user wallet: error: %+v, fund value: %v, userId: %v, leagueId: %v", err, value, userId, leagueId)
	}

	return updatedWalletValue, nil
}

func (u *UserRepo) incrementWalletFund(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (int64, error) {
	return u.redisClient.HIncrBy(
		context.Request().Context(),
		generateUserWalletRedisKey(userId),
		leagueId.String(),
		value,
	).Result()
}

func (u *UserRepo) GetUserIdFromSenderPsId(context echo.Context, senderPsId string) (uuid.UUID, error) {
	userIdString, err := u.redisClient.Get(
		context.Request().Context(),
		generateSenderPsIdToUserIdRedisKey(senderPsId),
	).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get userId from senderPsId: %v", senderPsId)
	}

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse userId: %v", userId)
	}

	return userId, nil
}

func (u *UserRepo) SetSenderPsIdToUserIdRelationship(context echo.Context, senderPsId string, userId uuid.UUID) error {
	_, err := u.redisClient.Set(
		context.Request().Context(),
		generateSenderPsIdToUserIdRedisKey(senderPsId),
		userId.String(),
		0,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to set senderPsId to userId relationship: senderPsId: %v, userId: %v", senderPsId, userId)
	}

	return nil
}
