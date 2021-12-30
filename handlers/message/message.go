package message

import (
	"bytes"
	"encoding/json"
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
		return context.JSON(http.StatusBadRequest, fmt.Errorf("missing mode or token from Messenger webhook"))
	}

	// Checks the mode and token sent is correct
	if mode == "subscribe" && token == secrets.MESSENGER_WEBHOOK_VERIFICATION_TOKEN {
		fmt.Println("WEBHOOK_VERIFIED")

		// Responds with the challenge token from the request
		return context.String(http.StatusOK, challenge)
	}

	// Responds with '403 Forbidden' if verify tokens do not match
	return context.JSON(http.StatusForbidden, fmt.Errorf("verify token does not match: requested token: %v", token))
}

func (m *MessageHandler) ProcessMessengerWebhook(context echo.Context) error {
	var body messenger_entities.WebhookBody

	err := json.NewDecoder(context.Request().Body).Decode(&body)
	if err != nil {
		return context.JSON(http.StatusBadRequest, fmt.Errorf("failed to decode body"))
	}

	// Returns a '404 Not Found' if event is not from a page subscription
	if body.Object != "page" {
		return context.JSON(http.StatusNotFound, fmt.Errorf("failed to find page event"))
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
		fmt.Println("webhook event processed")
	}

	if err != nil {
		context.Logger().Errorf("webhook processing error: %+v", err.Error())
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	// Returns a '200 OK' response to all requests
	return context.String(http.StatusOK, "EVENT_RECEIVED")
}

func (m *MessageHandler) HandleMessengerWebhookMessage(context echo.Context, senderPsId string, event messenger_entities.WebhookMessage) error {
	// Grab userId from the senderPsId
	userId, err := m.userService.GetUserIdFromSenderPsId(context, senderPsId)
	if err != nil {
		return context.JSON(http.StatusNotFound, fmt.Errorf("failed to find user from senderPsId"))
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

func (m *MessageHandler) SendMessage(context echo.Context) error {
	auctionId, err := m.auctionService.GetCurrentAuctionIdByLeagueId(context, constants.LEAGUE_ID)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	err = m.SendPlayersBidTemplateEvents(context, auctionId)
	if err != nil {
		context.Logger().Errorf("sending auction to users error: %+v", err.Error())
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(http.StatusOK, "ok")
}

func (m *MessageHandler) SendPlayersBidTemplateEvents(context echo.Context, auctionId uuid.UUID) error {
	sendEvents, err := m.auctionService.CreateBidsForAuction(context, auctionId)
	if err != nil {
		return context.JSON(http.StatusNotFound, err.Error())
	}

	for _, sendEvent := range sendEvents {
		sentEventJSON, _ := json.Marshal(sendEvent)

		postURL := fmt.Sprintf("https://graph.facebook.com/v12.0/me/messages?access_token=%v", secrets.MESSENGER_ACCESS_TOKEN)
		resp, err := http.Post(postURL, "application/json", bytes.NewBuffer(sentEventJSON))

		context.Logger().Debugf("response: %+v, error: %+v", resp, err)

		if err != nil {
			// handle error
			return context.JSON(http.StatusNotFound, err.Error())
		}

		defer resp.Body.Close()
	}

	return nil
}
