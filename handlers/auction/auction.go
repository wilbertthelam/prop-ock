package auction

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/constants"
	"github.com/wilbertthelam/prop-ock/entities"
	messenger_entities "github.com/wilbertthelam/prop-ock/entities/messenger"
	auction_service "github.com/wilbertthelam/prop-ock/services/auction"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
	"github.com/wilbertthelam/prop-ock/utils"
)

type AuctionHandler struct {
	auctionService *auction_service.AuctionService
	userService    *user_service.UserService
}

func New(
	auctionService *auction_service.AuctionService,
	userService *user_service.UserService,
) *AuctionHandler {
	return &AuctionHandler{
		auctionService,
		userService,
	}
}

func (a *AuctionHandler) GetBid(context echo.Context) error {
	params := context.QueryParams()

	auctionId, err := uuid.Parse(params.Get("auction_id"))
	if err != nil {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "failed to get bid params",
			Args: []interface{}{
				"auctionId", context.QueryParam("auction_id"),
			},
			Err: err,
		})
		return utils.JSONError(context, newErr)
	}

	// Make sure the sender has a userId
	// Grab userId from the senderPsId
	userId, err := a.userService.GetUserIdFromSenderPsId(context, params.Get("sender_ps_id"))
	if err != nil {
		return utils.JSONError(context, err)
	}

	bid, err := a.auctionService.GetBid(context, auctionId, userId, params.Get("player_id"))

	return context.JSON(http.StatusOK, bid)
}

func (a *AuctionHandler) MakeBid(context echo.Context) error {
	var body messenger_entities.WebhookBidPostBody

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "failed to decode make bid body",
			Err:     err,
		})
		return utils.JSONError(context, newErr)
	}

	// Make sure the sender has a userId
	// Grab userId from the senderPsId
	userId, err := a.userService.GetUserIdFromSenderPsId(context, body.SenderPsId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	auctionId, err := uuid.Parse(body.AuctionId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	playerId := body.PlayerId
	bid := body.Bid

	err = a.auctionService.MakeBid(context, auctionId, userId, playerId, bid)
	if err != nil {
		return utils.JSONError(context, err)
	}

	return context.JSON(http.StatusOK, "make bid successful")
}

func (a *AuctionHandler) CancelBid(context echo.Context) error {
	var body messenger_entities.WebhookBidPostBody

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "failed to decode cancel bid body",
			Err:     err,
		})
		return utils.JSONError(context, newErr)
	}

	// Make sure the sender has a userId
	// Grab userId from the senderPsId
	userId, err := a.userService.GetUserIdFromSenderPsId(context, body.SenderPsId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	auctionId, err := uuid.Parse(body.AuctionId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	playerId := body.PlayerId

	err = a.auctionService.CancelBid(context, auctionId, userId, playerId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	return context.JSON(http.StatusOK, "cancel bid successful")
}

func (a *AuctionHandler) CreateAuction(context echo.Context) error {
	leagueId := constants.LEAGUE_ID
	auctionId := uuid.New()

	auction, err := a.auctionService.CreateAuction(
		context,
		auctionId,
		leagueId,
		time.Now().UnixMilli(),
		time.Now().Add(time.Duration(10)*time.Minute).UnixMilli(),
	)
	if err != nil {
		return utils.JSONError(context, err)
	}

	// For now, start auction when we create it
	err = a.auctionService.StartAuction(context, auction.Id)
	if err != nil {
		return utils.JSONError(context, err)
	}

	return context.JSON(http.StatusOK, "created auction successful")
}

// Close auction stops the auction and prevents bids from coming in
func (a *AuctionHandler) StopAuction(context echo.Context) error {
	var body entities.Auction

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "failed to decode stop auction body",
			Err:     err,
		})
		return utils.JSONError(context, newErr)
	}

	auctionId := body.Id

	err = a.auctionService.StopAuction(context, auctionId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	return context.JSON(http.StatusOK, "stopping auction successful")
}

func (a *AuctionHandler) ProcessAuction(context echo.Context) error {
	var body entities.Auction

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		context.Logger().Error(err)
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "failed to close auction body",
			Err:     err,
		})
		return utils.JSONError(context, newErr)
	}

	auctionId := body.Id

	err = a.auctionService.ProcessAuction(context, auctionId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	return context.JSON(http.StatusOK, "processing auction successful")
}

func (a *AuctionHandler) GetCurrentAuctionForLeague(context echo.Context) error {
	leagueId, err := uuid.Parse(context.QueryParam("league_id"))
	if err != nil {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "failed to get current auction for league params",
			Err:     err,
		})
		return utils.JSONError(context, newErr)
	}

	activeAuctionId, err := a.auctionService.GetCurrentAuctionIdByLeagueId(context, leagueId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	auction, err := a.auctionService.GetAuctionByAuctionId(context, activeAuctionId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	return context.JSON(http.StatusOK, auction)
}

func (a *AuctionHandler) GetAuctionResults(context echo.Context) error {
	auctionId, err := uuid.Parse(context.QueryParam("auction_id"))
	if err != nil {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "failed to get auction params",
			Err:     err,
		})
		return utils.JSONError(context, newErr)
	}

	auctionResults, err := a.auctionService.GetAuctionResults(context, auctionId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	return context.JSON(http.StatusOK, auctionResults)
}
