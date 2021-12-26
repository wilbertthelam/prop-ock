package user_service

import (
	"fmt"

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

func (u *UserService) CreateUser(context echo.Context, userId uuid.UUID, name string) error {
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

func (u *UserService) AddUserToLeague(context echo.Context, userId uuid.UUID, leagueId uuid.UUID) error {
	// Verify user isn't already part of the league
	isUserInLeague, err := u.userRepo.IsUserMemberOfLeague(context, userId, leagueId)
	if err != nil || isUserInLeague {
		return err
	}

	return u.userRepo.AddUserToLeague(context, userId, leagueId)
}

func (u *UserService) RemoveUserFromLeague(context echo.Context, userId uuid.UUID, leagueId uuid.UUID) error {
	return nil
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
	return u.GetUserIdFromSenderPsId(context, senderPsId)
}

func (u *UserService) SetSenderPsIdToUserIdRelationship(context echo.Context, senderPsId string, userId uuid.UUID) error {
	return u.SetSenderPsIdToUserIdRelationship(context, senderPsId, userId)
}
