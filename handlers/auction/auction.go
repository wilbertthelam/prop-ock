package auction

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/constants"
	"github.com/wilbertthelam/prop-ock/entities"
	messenger_entities "github.com/wilbertthelam/prop-ock/entities/messenger"
	auction_service "github.com/wilbertthelam/prop-ock/services/auction"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
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

func (a *AuctionHandler) MakeBid(context echo.Context) error {
	var body messenger_entities.WebhookBidPostBody

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to decode bid body: %+v", err.Error()).Error())
	}

	// Validate all the post body fields

	// Make sure the sender has a userId
	// Grab userId from the senderPsId
	userId, err := a.userService.GetUserIdFromSenderPsId(context, body.SenderPsId)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to find user from senderPsId: %v", body.SenderPsId).Error())
	}

	auctionId, err := uuid.Parse(body.AuctionId)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to parse auctionId: %v", auctionId).Error())
	}

	playerId := body.PlayerId
	bid := body.Bid

	err = a.auctionService.MakeBid(context, auctionId, userId, playerId, bid)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to make bid: err: %+v, auctionId: %v, userId: %v, playerId: %v, bid: %v", err, auctionId, userId, playerId, bid).Error())
	}

	return context.JSON(http.StatusOK, "make bid successful")
}

func (a *AuctionHandler) CancelBid(context echo.Context) error {
	var body messenger_entities.WebhookBidPostBody

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to decode bid body: %+v", err.Error()).Error())
	}

	// Make sure the sender has a userId
	// Grab userId from the senderPsId
	userId, err := a.userService.GetUserIdFromSenderPsId(context, body.SenderPsId)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to find user from senderPsId: %v", body.SenderPsId).Error())
	}

	auctionId, err := uuid.Parse(body.AuctionId)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to parse auctionId: %v", auctionId).Error())
	}

	playerId := body.PlayerId

	err = a.auctionService.CancelBid(context, auctionId, userId, playerId)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to cancel bid: auctionId: %v, userId: %v, playerId: %v", auctionId, userId, playerId).Error())
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
		fmt.Printf("error: failed to create auction %+v", err)
		return context.JSON(http.StatusNotFound, err.Error())
	}

	// For now, start auction when we create it
	err = a.auctionService.StartAuction(context, auction.Id)
	if err != nil {
		fmt.Printf("error: failed to start auction %+v", err)
		return context.JSON(http.StatusNotFound, err.Error())
	}

	return nil
}

// Close auction stops the auction and prevents bids from coming in
func (a *AuctionHandler) StopAuction(context echo.Context) error {
	var body entities.Auction

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to decode close auction body: %+v", err.Error()).Error())
	}

	auctionId := body.Id

	err = a.auctionService.StopAuction(context, auctionId)
	if err != nil {
		fmt.Printf("error: failed to close auction %+v", err)
		return context.JSON(http.StatusNotFound, err.Error())
	}

	return nil
}

func (a *AuctionHandler) ProcessAuction(context echo.Context) error {
	var body entities.Auction

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to decode close auction body: %+v", err.Error()).Error())
	}

	auctionId := body.Id

	err = a.auctionService.ProcessAuction(context, auctionId)
	if err != nil {
		fmt.Printf("error: failed to process auction %+v", err)
		return context.JSON(http.StatusNotFound, err.Error())
	}

	return nil
}
