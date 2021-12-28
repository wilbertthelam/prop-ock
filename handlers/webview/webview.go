package webview

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	messenger_entities "github.com/wilbertthelam/prop-ock/entities/messenger"
	auction_service "github.com/wilbertthelam/prop-ock/services/auction"
	player_service "github.com/wilbertthelam/prop-ock/services/player"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
)

type WebviewHandler struct {
	playerService  *player_service.PlayerService
	auctionService *auction_service.AuctionService
	userService    *user_service.UserService
}

func New(
	playerService *player_service.PlayerService,
	auctionService *auction_service.AuctionService,
	userService *user_service.UserService,
) *WebviewHandler {
	return &WebviewHandler{
		playerService,
		auctionService,
		userService,
	}
}

func GetName() string {
	return "webview"
}

func (w *WebviewHandler) GetPlayer(context echo.Context) error {
	player, err := w.playerService.GetPlayerByPlayerId(context, context.QueryParam("playerId"))
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}
	return context.JSON(http.StatusOK, player)
}

func (w *WebviewHandler) MakeBid(context echo.Context) error {
	var body messenger_entities.WebhookBidPostBody

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to decode bid body: %+v", err.Error()).Error())
	}

	// Validate all the post body fields

	// Make sure the sender has a userId
	// Grab userId from the senderPsId
	userId, err := w.userService.GetUserIdFromSenderPsId(context, body.SenderPsId)
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

	err = w.auctionService.MakeBid(context, auctionId, userId, playerId, bid)
	if err != nil {
		context.Logger().Error(err)
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to make bid: auctionId: %v, userId: %v, playerId: %v, bid: %v", auctionId, userId, playerId, bid).Error())
	}

	return context.JSON(http.StatusOK, "bid successful")
}
