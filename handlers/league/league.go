package league

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/constants"
	league_service "github.com/wilbertthelam/prop-ock/services/league"
)

type LeagueHandler struct {
	leagueService *league_service.LeagueService
}

func New(
	leagueService *league_service.LeagueService,
) *LeagueHandler {
	return &LeagueHandler{
		leagueService,
	}
}

func (l *LeagueHandler) GetLeagueCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, "ok")
}

func (l *LeagueHandler) CreateLeague(context echo.Context) error {
	leagueId := constants.LEAGUE_ID
	err := l.leagueService.CreateLeague(context, leagueId, "Field of GGreams")
	if err != nil {
		fmt.Printf("error: failed to create league %+v", err)
		return context.JSON(http.StatusNotFound, err.Error())
	}

	return nil
}
