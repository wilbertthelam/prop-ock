package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

func GetName() string {
	return "message"
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

		break
	}

	return nil
}

func (m *MessageHandler) HandleMessengerWebhookRead(context echo.Context, senderPsId string, event messenger_entities.WebhookRead) error {
	return nil
}

func (m *MessageHandler) SendMessage(context echo.Context) error {
	auctionId, err := m.auctionService.GetAuctionIdByLeagueId(context, constants.LEAGUE_ID)
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

func (m *MessageHandler) CreateLeague(context echo.Context) error {
	leagueId := constants.LEAGUE_ID
	err := m.leagueService.CreateLeague(context, leagueId, "Field of GGreams")
	if err != nil {
		fmt.Printf("error: failed to create league %+v", err)
		return context.JSON(http.StatusNotFound, err.Error())
	}

	return nil
}

func (m *MessageHandler) CreateAuction(context echo.Context) error {
	leagueId := constants.LEAGUE_ID
	auctionId := uuid.New()

	auction, err := m.auctionService.CreateAuction(
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

	err = m.auctionService.StartAuction(context, auction.Id)
	if err != nil {
		fmt.Printf("error: failed to start auction %+v", err)
		return context.JSON(http.StatusNotFound, err.Error())
	}

	return nil

	// // err = m.auctionService.CloseAuction(context, auction.Id)
	// // if err != nil {
	// // 	fmt.Printf("error: failed to close auction %+v", err)
	// // 	return context.JSON(http.StatusNotFound, err.Error())
	// // }

	// // err = m.auctionService.ArchiveAuction(context, auction.Id)
	// // if err != nil {
	// // 	fmt.Printf("error: failed to archive auction %+v", err)
	// // 	return context.JSON(http.StatusNotFound, err.Error())
	// // }

	// // return context.JSON(http.StatusOK, fmt.Sprintf("latest message - auctionId: %v", auction.Id.String()))
	// leagueId := uuid.MustParse("894098e8-8cfe-4c92-9e32-332aac801899")
	// user1Id := uuid.MustParse("5ce0beb6-e12b-42c0-adb4-4153bff08eb9")
	// user2Id := uuid.MustParse("242e7749-8816-4053-9fdd-3292e4122fed")

	// err = m.userService.CreateUser(context, user1Id, "Fred Johnson")
	// if err != nil {
	// 	fmt.Printf("error: failed to create user %+v", err)
	// 	return context.JSON(http.StatusNotFound, err.Error())
	// }

	// err = m.userService.CreateUser(context, user2Id, "Bobbi Draper")
	// if err != nil {
	// 	fmt.Printf("error: failed to create user %+v", err)
	// 	return context.JSON(http.StatusNotFound, err.Error())
	// }

	// user1, err := m.userService.GetUserByUserId(context, user1Id)
	// if err != nil {
	// 	fmt.Printf("error: failed to get user %+v", err)
	// 	return context.JSON(http.StatusNotFound, err.Error())
	// }

	// _, err = m.userService.GetUserByUserId(context, user2Id)
	// if err != nil {
	// 	fmt.Printf("error: failed to get user %+v", err)
	// 	return context.JSON(http.StatusNotFound, err.Error())
	// }

	// err = m.userService.AddUserToLeague(context, user1Id, leagueId)
	// if err != nil {
	// 	fmt.Printf("error: failed to add user to league %+v", err)
	// 	return context.JSON(http.StatusNotFound, err.Error())
	// }

	// err = m.userService.AddUserToLeague(context, user2Id, leagueId)
	// if err != nil {
	// 	fmt.Printf("error: failed to add user to league %+v", err)
	// 	return context.JSON(http.StatusNotFound, err.Error())
	// }

	// return context.JSON(http.StatusOK, user1)
}

// func (m *MessageHandler) SendMessage(context echo.Context) error {
// 	// Message state machine

// }

// func (m *MessageHandler) SendMessage(context echo.Context) error {
// 	leagueId := uuid.MustParse("894098e8-8cfe-4c92-9e32-332aac801899")
// 	user1Id := uuid.MustParse("5ce0beb6-e12b-42c0-adb4-4153bff08eb9")
// 	user2Id := uuid.MustParse("242e7749-8816-4053-9fdd-3292e4122fed")
// 	playerId := "12345-julio-rodriguez"
// 	auctionId := uuid.MustParse("c40d070c-931e-44ae-820b-46d595d9af6e")

// 	updatedWallet, err := m.userService.AddFundsToUserWallet(context, user1Id, leagueId, 100)
// 	if err != nil {
// 		return context.JSON(http.StatusNotFound, err.Error())
// 	}

// 	fmt.Println(updatedWallet)

// 	updatedWallet, err = m.userService.RemoveFundsFromUserWallet(context, user1Id, leagueId, 25)
// 	if err != nil {
// 		return context.JSON(http.StatusNotFound, err.Error())
// 	}

// 	fmt.Println(updatedWallet)

// 	updatedWallet, err = m.userService.AddFundsToUserWallet(context, user1Id, leagueId, 50)
// 	if err != nil {
// 		return context.JSON(http.StatusNotFound, err.Error())
// 	}

// 	fmt.Println(updatedWallet)

// 	updatedWallet, err = m.userService.AddFundsToUserWallet(context, user2Id, leagueId, 500)
// 	if err != nil {
// 		return context.JSON(http.StatusNotFound, err.Error())
// 	}

// 	err = m.auctionService.MakeBid(context, auctionId, user1Id, playerId, 50)
// 	if err != nil {
// 		fmt.Printf("error: failed to make bid for user %+v: ", err)
// 		return context.JSON(http.StatusNotFound, err.Error())
// 	}

// 	err = m.auctionService.MakeBid(context, auctionId, user2Id, playerId, 100)
// 	if err != nil {
// 		fmt.Printf("error: failed to make bid for user %+v: ", err)
// 		return context.JSON(http.StatusNotFound, err.Error())
// 	}

// 	err = m.auctionService.MakeBid(context, auctionId, user2Id, playerId, 150)
// 	if err != nil {
// 		fmt.Printf("error: failed to make bid for user %+v: ", err)
// 		return context.JSON(http.StatusNotFound, err.Error())
// 	}

// 	err = m.auctionService.CancelBid(context, auctionId, user1Id, playerId)
// 	if err != nil {
// 		fmt.Printf("error: failed to cancel bid for user %+v: ", err)
// 		return context.JSON(http.StatusNotFound, err.Error())
// 	}

// 	return context.JSON(http.StatusOK, updatedWallet)
// }
