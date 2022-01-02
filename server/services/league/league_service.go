package league_service

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
	league_repo "github.com/wilbertthelam/prop-ock/repos/league"
	"github.com/wilbertthelam/prop-ock/utils"
)

type LeagueService struct {
	leagueRepo *league_repo.LeagueRepo
}

func New(leagueRepo *league_repo.LeagueRepo) *LeagueService {
	return &LeagueService{
		leagueRepo,
	}
}

func (l *LeagueService) GetLeagueByLeagueId(context echo.Context, leagueId uuid.UUID) (entities.League, error) {
	return l.leagueRepo.GetLeagueByLeagueId(context, leagueId)
}

func (l *LeagueService) GetMembersInLeague(context echo.Context, leagueId uuid.UUID) ([]uuid.UUID, error) {
	return l.leagueRepo.GetMembersInLeague(context, leagueId)
}

func (l *LeagueService) IsUserInLeague(context echo.Context, userId uuid.UUID, leagueId uuid.UUID) (bool, error) {
	members, err := l.GetMembersInLeague(context, leagueId)
	if err != nil {
		return false, err
	}

	for _, memberId := range members {
		if memberId == userId {
			return true, nil
		}
	}

	return false, nil
}

func (l *LeagueService) AddUserToLeague(context echo.Context, userId uuid.UUID, leagueId uuid.UUID) error {
	// Verify user isn't already part of the league
	isUserInLeague, err := l.leagueRepo.IsUserMemberOfLeague(context, userId, leagueId)
	if err != nil {
		return err
	}

	if isUserInLeague {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "user already a member of this league",
			Args: []interface{}{
				"userId", userId.String(),
				"leagueId", leagueId.String(),
			},
			Err: nil,
		})
	}

	return l.leagueRepo.AddUserToLeague(context, userId, leagueId)
}

func (l *LeagueService) RemoveUserFromLeague(context echo.Context, userId uuid.UUID, leagueId uuid.UUID) error {
	return nil
}

func (l *LeagueService) CreateLeague(context echo.Context, leagueId uuid.UUID, name string) error {
	// Verify league isn't already created
	league, err := l.GetLeagueByLeagueId(context, leagueId)
	if err != nil {
		return err
	}

	if league.Id != uuid.Nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "league is already created",
			Args: []interface{}{
				"leagueId", leagueId.String(),
				"leagueName", league.Name,
			},
			Err: nil,
		})
	}

	league = entities.League{
		Id:   leagueId,
		Name: name,
	}

	return l.leagueRepo.CreateLeague(context, leagueId, league)
}
