package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/constants"
	"github.com/wilbertthelam/prop-ock/entities"
	messenger_entities "github.com/wilbertthelam/prop-ock/entities/messenger"
	"github.com/wilbertthelam/prop-ock/secrets"
	auction_service "github.com/wilbertthelam/prop-ock/services/auction"
	callups_service "github.com/wilbertthelam/prop-ock/services/callups"
	league_service "github.com/wilbertthelam/prop-ock/services/league"
	message_service "github.com/wilbertthelam/prop-ock/services/message"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
	"github.com/wilbertthelam/prop-ock/utils"
)

type MessageHandler struct {
	auctionService *auction_service.AuctionService
	callupsService *callups_service.CallupsService
	userService    *user_service.UserService
	leagueService  *league_service.LeagueService
	messageService *message_service.MessageService
}

func New(
	auctionService *auction_service.AuctionService,
	callupsService *callups_service.CallupsService,
	userService *user_service.UserService,
	leagueService *league_service.LeagueService,
	messageService *message_service.MessageService,
) *MessageHandler {
	return &MessageHandler{
		auctionService,
		callupsService,
		userService,
		leagueService,
		messageService,
	}
}

func (m *MessageHandler) VerifyMessengerWebhook(context echo.Context) error {
	mode := context.QueryParam("hub.mode")
	token := context.QueryParam("hub.verify_token")
	challenge := context.QueryParam("hub.challenge")

	// Checks if a token and mode is in the query string of the request
	if mode == "" || token == "" {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "missing mode or token from Messenger webhook",
			Args: []interface{}{
				"mode", mode,
				"token", token,
			},
			Err: nil,
		})
		return utils.JSONError(context, newErr)
	}

	// Checks the mode and token sent is correct
	if mode == "subscribe" && token == secrets.MESSENGER_WEBHOOK_VERIFICATION_TOKEN {
		context.Logger().Info("WEBHOOK_VERIFIED")

		// Responds with the challenge token from the request
		return context.String(http.StatusOK, challenge)
	}

	// Responds with '403 Forbidden' if verify tokens do not match
	return context.String(http.StatusForbidden, "tokens do not match")
}

func (m *MessageHandler) ProcessMessengerWebhook(context echo.Context) error {
	var body messenger_entities.WebhookBody

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "failed to decode messenger webhook body",
			Err:     err,
		})
		return utils.JSONError(context, newErr)
	}

	// Returns a '404 Not Found' if event is not from a page subscription
	if body.Object != "page" {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusNotFound,
			Message: "failed to find page event",
			Err:     nil,
		})
		return utils.JSONError(context, newErr)
	}

	// Iterates over each entry - there may be multiple if batched
	for _, entry := range body.Entry {
		// Gets the message. entry.messaging is an array, but
		// will only ever contain one message, so we get index 0
		webhookEvent := entry.Messaging[0]

		context.Logger().Infof("webhook event: %+v", webhookEvent)

		// Process webhook event here
		// Get the sender PSID
		senderPsId := webhookEvent.Sender.Id
		context.Logger().Infof("senderPsId: %v", senderPsId)

		// Check if the event is a message or postback or read and
		// pass the event to the appropriate handler function
		if (webhookEvent.Message != messenger_entities.WebhookMessage{}) {
			err = m.HandleMessengerWebhookMessage(context, senderPsId, webhookEvent.Message)
		} else if (webhookEvent.Postback != messenger_entities.WebhookPostback{}) {
			err = m.HandleMessengerWebhookPostback(context, senderPsId, webhookEvent.Postback)
		} else if (webhookEvent.Read != messenger_entities.WebhookRead{}) {
			err = m.HandleMessengerWebhookRead(context, senderPsId, webhookEvent.Read)
		}

		// Webhook event processed
		context.Logger().Info("webhook event processed")
	}

	if err != nil {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "webhook processing error",
			Err:     err,
		})
		return utils.JSONError(context, newErr)
	}

	// Returns a '200 OK' response to all requests
	return context.String(http.StatusOK, "EVENT_RECEIVED")
}

func (m *MessageHandler) HandleMessengerWebhookMessage(context echo.Context, senderPsId string, event messenger_entities.WebhookMessage) error {
	// Grab userId from the senderPsId
	userId, err := m.userService.GetUserIdFromSenderPsId(context, senderPsId)
	if err != nil {
		return err
	}

	return m.messageService.SendAction(context, entities.ACTION_SEND_MESSAGE, userId, event)
}

func (m *MessageHandler) HandleMessengerWebhookPostback(context echo.Context, senderPsId string, event messenger_entities.WebhookPostback) error {
	// Chest Postback type

	// New user initialization type
	switch event.Payload {
	case "user_joined":
		// Initialize new userId
		userId := uuid.New()
		leagueId := constants.LEAGUE_ID

		err := m.userService.InitializeUser(context, userId, senderPsId, "[add-name]")
		if err != nil {
			// TODO: figure out what to do if initialization fails
			return err
		}

		// Join the league
		err = m.leagueService.AddUserToLeague(context, userId, leagueId)
		if err != nil {
			return err
		}

		// Add starting funds to their wallet
		_, err = m.userService.AddFundsToUserWallet(context, userId, leagueId, constants.STARTING_WALLET_AMOUNT)
		if err != nil {
			return err
		}

		break
	}

	return nil
}

func (m *MessageHandler) HandleMessengerWebhookRead(context echo.Context, senderPsId string, event messenger_entities.WebhookRead) error {
	return nil
}

func (m *MessageHandler) SendWinningBids(context echo.Context) error {
	auctionId, err := m.auctionService.GetCurrentAuctionIdByLeagueId(context, constants.LEAGUE_ID)
	if err != nil {
		return utils.JSONError(context, err)
	}

	auctionResults, err := m.auctionService.GetAuctionResults(context, auctionId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	// For each winning bid, create a success response for it
	playerEvents := make([]messenger_entities.SendEvent, 0)
	for _, winningBids := range auctionResults {
		if len(winningBids) > 1 {
			// TODO: handle tie case
		}

		winningBid := winningBids[0]

		playerEvent, err := m.messageService.CreateWinningBidForPlayerEvent(context, winningBid)
		if err != nil {
			return utils.JSONError(context, err)
		}

		playerEvent, err = m.attachSenderToEvent(context, winningBid.UserId, playerEvent)
		if err != nil {
			return utils.JSONError(context, err)
		}

		playerEvent = m.attachConnectionTagToEvent(context, playerEvent)

		playerEvents = append(playerEvents, playerEvent)
	}

	// Once all events are generated, send them out
	errListMap := m.sendEvents(context, playerEvents)
	errList := make([]interface{}, len(errListMap))
	for index, err := range errList {
		errList[index] = err
	}
	if len(errListMap) > 0 {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to send winning bid events to users on messenger",
			Args:    errList,
			Err:     errors.New("error list"),
		})
		return utils.JSONError(context, newErr)
	}

	return context.JSON(http.StatusOK, "ok")
}

func (m *MessageHandler) attachConnectionTagToEvent(context echo.Context, event messenger_entities.SendEvent) messenger_entities.SendEvent {
	// Attached required tags to the event to make sure that we can keep
	// sending the user messages after 24 hours
	event.Tag = constants.CONFIRM_TAG_UPDATE
	return event
}

func (m *MessageHandler) SendPlayersForBidding(context echo.Context) error {
	auctionId, err := m.auctionService.GetCurrentAuctionIdByLeagueId(context, constants.LEAGUE_ID)
	if err != nil {
		return utils.JSONError(context, err)
	}

	sendEvents, err := m.messageService.CreateBidsForAuction(context, auctionId)
	if err != nil {
		return utils.JSONError(context, err)
	}

	for index, sendEvent := range sendEvents {
		sendEvent = m.attachConnectionTagToEvent(context, sendEvent)
		sendEvents[index] = sendEvent
	}

	// Once all events are generated, send them out
	errListMap := m.sendEvents(context, sendEvents)
	errList := make([]interface{}, len(errListMap))
	for index, err := range errList {
		errList[index] = err
	}
	if len(errListMap) > 0 {
		newErr := utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to send auction bid events to users on messenger",
			Args:    errList,
			Err:     errors.New("error list"),
		})
		return utils.JSONError(context, newErr)
	}

	return context.JSON(http.StatusOK, "ok")
}

func (m *MessageHandler) attachSenderToEvent(context echo.Context, userId uuid.UUID, event messenger_entities.SendEvent) (messenger_entities.SendEvent, error) {
	senderPsId, err := m.userService.GetSenderPsIdFromUserId(context, userId)
	if err != nil {
		return messenger_entities.SendEvent{}, err
	}

	// Attach intended sender the event should be directed towards
	event.Recipient = messenger_entities.Id{
		Id: senderPsId,
	}

	return event, nil
}

func (m *MessageHandler) sendEvents(context echo.Context, sendEvents []messenger_entities.SendEvent) map[string]error {
	postURL := fmt.Sprintf("https://graph.facebook.com/v12.0/me/messages?access_token=%v", secrets.MESSENGER_ACCESS_TOKEN)

	errors := make(map[string]error)
	for _, sendEvent := range sendEvents {
		sendEventJSON, _ := json.Marshal(sendEvent)

		rawResp, httpErr := http.Post(postURL, "application/json", bytes.NewBuffer(sendEventJSON))
		context.Logger().Infof("response: %+v, error: %+v", rawResp, httpErr)

		if httpErr != nil {
			newHttpErr := utils.NewError(utils.ErrorParams{
				Code:    http.StatusBadRequest,
				Message: "failed to post send event request",
				Args: []interface{}{
					"sendEventJson", string(sendEventJSON),
				},
				Err: httpErr,
			})
			errors[sendEvent.Recipient.Id] = newHttpErr
			continue
		}

		var resp messenger_entities.SendEventResponse
		decodeErr := json.NewDecoder(rawResp.Body).Decode(&resp)
		if decodeErr != nil {
			newDecodeErr := utils.NewError(utils.ErrorParams{
				Code:    http.StatusBadRequest,
				Message: "failed to decode send event request",
				Err:     httpErr,
			})
			errors[sendEvent.Recipient.Id] = newDecodeErr
			continue
		}

		// If the Messenger SendAPI returns an error, give us a heads up
		if resp.Error.Code > 0 {
			respErr := utils.NewError(utils.ErrorParams{
				Code:    http.StatusBadRequest,
				Message: "failed to post send event request",
				Args: []interface{}{
					"error", resp.Error,
				},
			})
			errors[sendEvent.Recipient.Id] = respErr
			continue
		}

		defer rawResp.Body.Close()
	}

	return errors
}
