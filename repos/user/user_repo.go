package user_repo

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	redis_client "github.com/wilbertthelam/prop-ock/db"
	"github.com/wilbertthelam/prop-ock/entities"
	"github.com/wilbertthelam/prop-ock/utils"
)

type UserRepo struct {
	redisClient *redis.Client
}

func New(redisClient *redis.Client) *UserRepo {
	return &UserRepo{
		redisClient,
	}
}

func generateUserRedisKey(userId uuid.UUID) string {
	return fmt.Sprintf("user:user_id:%v", userId.String())
}

func generateUserWalletRedisKey(userId uuid.UUID) string {
	return fmt.Sprintf("wallet:user_id:%v", userId.String())
}

func generateSenderPsIdToUserIdRedisKey(senderPsId string) string {
	return fmt.Sprintf("relationship:sender_ps_id_to_user_id:%v", senderPsId)
}

func generateUserIdToSenderPsIdRedisKey(userId uuid.UUID) string {
	return fmt.Sprintf("relationship:user_id_to_sender_ps_id:%v", userId.String())
}

func (u *UserRepo) GetUserByUserId(context echo.Context, userId uuid.UUID) (entities.User, error) {
	redisUser, err := u.redisClient.HGetAll(
		context.Request().Context(),
		generateUserRedisKey(userId),
	).Result()
	if err != nil {
		return entities.User{}, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to get user",
			Args: []interface{}{
				"userId", userId.String(),
			},
			Err: err,
		})
	}

	// If user is not found, then return an empty user
	if len(redisUser) == 0 {
		return entities.User{}, nil
	}

	user := entities.User{
		Id:   uuid.MustParse(redisUser["id"]),
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

func (u *UserRepo) updateUser(context echo.Context, userId uuid.UUID, keyValuePairs []string) error {
	_, err := redis_client.
		GetCmdable(context, u.redisClient).
		HSet(
			context.Request().Context(),
			generateUserRedisKey(userId),
			keyValuePairs,
		).Result()
	if err != nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to update user fields",
			Args: append(
				[]interface{}{"userId", userId.String()},
				utils.MapStringSliceToInterfaceSlice(keyValuePairs)...,
			),
			Err: err,
		})
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
		return nil, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to get user wallet player fields",
			Args: []interface{}{
				"userId", userId.String(),
			},
			Err: err,
		})
	}

	for key, value := range redisWallet {
		uuidKey, err := uuid.Parse(key)
		if err != nil {
			return nil, utils.NewError(utils.ErrorParams{
				Code:    http.StatusInternalServerError,
				Message: "failed to parse user wallet player leagueId key",
				Args: []interface{}{
					"userId", userId.String(),
					"key", key,
				},
				Err: err,
			})
		}

		int64Value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, utils.NewError(utils.ErrorParams{
				Code:    http.StatusInternalServerError,
				Message: "failed to parse user wallet player leagueId value",
				Args: []interface{}{
					"userId", userId.String(),
					"value", value,
				},
				Err: err,
			})
		}

		wallet[uuidKey] = int64Value
	}

	return wallet, nil
}

func (u *UserRepo) AddFundsToUserWallet(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (int64, error) {
	updatedWalletValue, err := u.incrementWalletFund(context, userId, leagueId, value)
	if err != nil {
		return 0, err
	}

	return updatedWalletValue, nil
}

func (u *UserRepo) RemoveFundsFromUserWallet(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (int64, error) {
	updatedWalletValue, err := u.incrementWalletFund(context, userId, leagueId, value*-1)
	if err != nil {
		return 0, err
	}

	return updatedWalletValue, nil
}

func (u *UserRepo) incrementWalletFund(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (int64, error) {
	updatedWalletValue, err := redis_client.
		GetCmdable(context, u.redisClient).
		HIncrBy(
			context.Request().Context(),
			generateUserWalletRedisKey(userId),
			leagueId.String(),
			value,
		).Result()
	if err != nil {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to adjust wallet fund",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: err,
		})
	}

	return updatedWalletValue, nil
}

func (u *UserRepo) GetUserIdFromSenderPsId(context echo.Context, senderPsId string) (uuid.UUID, error) {
	userIdString, err := u.redisClient.Get(
		context.Request().Context(),
		generateSenderPsIdToUserIdRedisKey(senderPsId),
	).Result()
	if err != nil {
		return uuid.Nil, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to get userId from senderPsId",
			Args: []interface{}{
				"senderPsId", senderPsId,
			},
			Err: err,
		})
	}

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		return uuid.Nil, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to parse userId from senderPsId",
			Args: []interface{}{
				"senderPsId", senderPsId,
			},
			Err: err,
		})
	}

	return userId, nil
}

func (u *UserRepo) GetSenderPsIdFromUserId(context echo.Context, userId uuid.UUID) (string, error) {
	senderPsId, err := u.redisClient.Get(
		context.Request().Context(),
		generateUserIdToSenderPsIdRedisKey(userId),
	).Result()
	if err != nil {
		return "", utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to get senderPsId from userId",
			Args: []interface{}{
				"userId", userId.String(),
			},
			Err: err,
		})
	}

	return senderPsId, nil
}

func (u *UserRepo) SetSenderPsIdToUserIdRelationship(context echo.Context, senderPsId string, userId uuid.UUID) error {
	_, err := redis_client.
		GetCmdable(context, u.redisClient).
		Set(
			context.Request().Context(),
			generateSenderPsIdToUserIdRedisKey(senderPsId),
			userId.String(),
			0,
		).Result()
	if err != nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to set senderPsId to userId relationship",
			Args: []interface{}{
				"senderPsId", senderPsId,
				"userId", userId.String(),
			},
			Err: err,
		})
	}

	return nil
}

func (u *UserRepo) SetUserIdToSenderPsIdRelationship(context echo.Context, senderPsId string, userId uuid.UUID) error {
	_, err := redis_client.
		GetCmdable(context, u.redisClient).
		Set(
			context.Request().Context(),
			generateUserIdToSenderPsIdRedisKey(userId),
			senderPsId,
			0,
		).Result()
	if err != nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to set userId to senderPsId relationship",
			Args: []interface{}{
				"userId", userId.String(),
				"senderPsId", senderPsId,
			},
			Err: err,
		})
	}

	return nil
}
