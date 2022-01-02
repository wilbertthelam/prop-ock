package user_service

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	redis_client "github.com/wilbertthelam/prop-ock/db"
	"github.com/wilbertthelam/prop-ock/entities"
	user_repo "github.com/wilbertthelam/prop-ock/repos/user"
	league_service "github.com/wilbertthelam/prop-ock/services/league"
	"github.com/wilbertthelam/prop-ock/utils"
)

type UserService struct {
	userRepo      *user_repo.UserRepo
	leagueService *league_service.LeagueService
	redisClient   *redis.Client
}

func New(
	userRepo *user_repo.UserRepo,
	leagueService *league_service.LeagueService,
	redisClient *redis.Client,
) *UserService {
	return &UserService{
		userRepo,
		leagueService,
		redisClient,
	}
}

func (u *UserService) GetUserByUserId(context echo.Context, userId uuid.UUID) (entities.User, error) {
	return u.userRepo.GetUserByUserId(context, userId)
}

func (u *UserService) InitializeUser(context echo.Context, userId uuid.UUID, senderPsId string, name string) error {
	// Check if account was already created for new user
	checkedUserId, err := u.GetUserIdFromSenderPsId(context, senderPsId)
	if err == nil {
		return err
	}

	if checkedUserId != uuid.Nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "userId already mapped to senderPsId",
			Args: []interface{}{
				"userId", userId.String(),
				"senderPsId", senderPsId,
			},
			Err: nil,
		})
	}

	// Start Redis transaction to create user and relationships
	err = redis_client.StartTransaction(
		context,
		u.redisClient,
		func() error {
			err = u.createUser(context, userId, "[add name]")
			if err != nil {
				return err
			}

			// Create senderPsId <> userId mappings in both directions
			err = u.SetUserIdToSenderPsIdRelationship(context, senderPsId, userId)
			if err != nil {
				return err
			}

			err = u.SetSenderPsIdToUserIdRelationship(context, senderPsId, userId)
			if err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserService) createUser(context echo.Context, userId uuid.UUID, name string) error {
	// Verify user isn't already created
	user, err := u.GetUserByUserId(context, userId)
	if err != nil {
		return err
	}

	if user.Id != uuid.Nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "user is already created",
			Args: []interface{}{
				"userId", userId.String(),
			},
			Err: nil,
		})
	}

	user = entities.User{
		Id:   userId,
		Name: name,
	}

	return u.userRepo.CreateUser(context, userId, user)
}

func (u *UserService) GetUserWallet(context echo.Context, userId uuid.UUID) (map[uuid.UUID]int64, error) {
	return u.userRepo.GetUserWallet(context, userId)
}

func (u *UserService) AddFundsToUserWallet(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (int64, error) {
	// Verify user exists
	user, err := u.GetUserByUserId(context, userId)
	if err != nil {
		return 0, err
	}

	if userId != user.Id {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "user does not exist",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: nil,
		})
	}

	// Verify league exists
	league, err := u.leagueService.GetLeagueByLeagueId(context, leagueId)
	if err != nil {
		return 0, err
	}

	if leagueId != league.Id {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "league does not exist",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: nil,
		})
	}

	// Verify user is in league
	isUserInLeague, err := u.leagueService.IsUserInLeague(context, userId, leagueId)
	if err != nil {
		return 0, err
	}

	if isUserInLeague == false {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusForbidden,
			Message: "user does not exist in this league",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: nil,
		})
	}

	// Verify value is positive
	if value < 0 {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "added fund must be a positive value",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: nil,
		})
	}

	return u.userRepo.AddFundsToUserWallet(context, userId, leagueId, value)
}

func (u *UserService) RemoveFundsFromUserWallet(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (int64, error) {
	// Verify user exists
	user, err := u.GetUserByUserId(context, userId)
	if err != nil {
		return 0, err
	}

	if userId != user.Id {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "user does not exist",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: nil,
		})
	}

	// Verify league exists
	league, err := u.leagueService.GetLeagueByLeagueId(context, leagueId)
	if err != nil {
		return 0, err
	}

	if leagueId != league.Id {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "league does not exist",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: nil,
		})
	}

	// Verify user is in league
	isUserInLeague, err := u.leagueService.IsUserInLeague(context, userId, leagueId)
	if err != nil {
		return 0, err
	}

	if isUserInLeague == false {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusForbidden,
			Message: "user does not exist in this league",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: nil,
		})
	}

	// Verify value is positive
	if value < 0 {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "removed fund must be a positive value",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: nil,
		})
	}

	// Verify the user has enough funds to remove
	hasEnoughFunds, err := u.ValidateUserHasEnoughFunds(context, userId, leagueId, value)
	if err != nil {
		return 0, err
	}

	if !hasEnoughFunds {
		return 0, utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "wallet does not have enough funds to remove value",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
				"value", fmt.Sprintf("%v", value),
			},
			Err: nil,
		})
	}

	return u.userRepo.RemoveFundsFromUserWallet(context, userId, leagueId, value)
}

func (u *UserService) ValidateUserHasEnoughFunds(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (bool, error) {
	wallet, err := u.GetUserWallet(context, userId)
	if err != nil {
		return false, err
	}

	return wallet[leagueId] >= value, nil
}

func (u *UserService) GetUserIdFromSenderPsId(context echo.Context, senderPsId string) (uuid.UUID, error) {
	return u.userRepo.GetUserIdFromSenderPsId(context, senderPsId)
}

func (u *UserService) SetSenderPsIdToUserIdRelationship(context echo.Context, senderPsId string, userId uuid.UUID) error {
	return u.userRepo.SetSenderPsIdToUserIdRelationship(context, senderPsId, userId)
}

func (u *UserService) GetSenderPsIdFromUserId(context echo.Context, userId uuid.UUID) (string, error) {
	return u.userRepo.GetSenderPsIdFromUserId(context, userId)
}

func (u *UserService) SetUserIdToSenderPsIdRelationship(context echo.Context, senderPsId string, userId uuid.UUID) error {
	return u.userRepo.SetUserIdToSenderPsIdRelationship(context, senderPsId, userId)
}
