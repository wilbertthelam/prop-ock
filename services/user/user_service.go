package user_service

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
	user_repo "github.com/wilbertthelam/prop-ock/repos/user"
)

type UserService struct {
	userRepo *user_repo.UserRepo
}

func New(userRepo *user_repo.UserRepo) *UserService {
	return &UserService{
		userRepo,
	}
}

func GetName() string {
	return "user_service"
}

func (u *UserService) GetUserByUserId(context echo.Context, userId uuid.UUID) (entities.User, error) {
	return u.userRepo.GetUserByUserId(context, userId)
}

func (u *UserService) InitializeUser(context echo.Context, userId uuid.UUID, senderPsId string, name string) error {
	// Check if account was already created for new user
	checkedUserId, err := u.GetUserIdFromSenderPsId(context, senderPsId)
	if err == nil || checkedUserId != uuid.Nil {
		return fmt.Errorf("user already mapped: senderPsId: %v, userId: %v, error: %+v", senderPsId, userId, err)
	}

	// TODO: Create Redis transaction here

	err = u.createUser(context, userId, "[add name]")
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
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
}

func (u *UserService) createUser(context echo.Context, userId uuid.UUID, name string) error {
	// Verify user isn't already created
	user, err := u.GetUserByUserId(context, userId)
	if err != nil {
		return err
	}

	if user.Id != uuid.Nil {
		return fmt.Errorf("user is already created: %v", userId)
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

	// Verify league exists

	// Verify value is positive
	if value < 0 {
		return 0, fmt.Errorf("added fund must be a positive value: %v", value)
	}

	return u.userRepo.AddFundsToUserWallet(context, userId, leagueId, value)
}

func (u *UserService) RemoveFundsFromUserWallet(context echo.Context, userId uuid.UUID, leagueId uuid.UUID, value int64) (int64, error) {
	// Verify user exists

	// Verify league exists

	// Verify value is positive
	if value < 0 {
		return 0, fmt.Errorf("removed fund must be a positive value: %v", value)
	}

	// Verify the user has enough funds to remove
	hasEnoughFunds, err := u.ValidateUserHasEnoughFunds(context, userId, leagueId, value)
	if err != nil {
		return 0, err
	}

	if !hasEnoughFunds {
		return 0, fmt.Errorf("wallet does not have enough funds to remove value: %v", value)
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
