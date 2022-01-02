package message_service

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/constants"
	"github.com/wilbertthelam/prop-ock/entities"
	messenger_entities "github.com/wilbertthelam/prop-ock/entities/messenger"
	auction_service "github.com/wilbertthelam/prop-ock/services/auction"
	league_service "github.com/wilbertthelam/prop-ock/services/league"
	player_service "github.com/wilbertthelam/prop-ock/services/player"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
)

type MessageService struct {
	auctionService *auction_service.AuctionService
	userService    *user_service.UserService
	playerService  *player_service.PlayerService
	leagueService  *league_service.LeagueService
	state          *State
}

type State struct {
	state map[uuid.UUID]entities.MessageState
	mutex sync.Mutex
}

func NewState() *State {
	return &State{
		state: make(map[uuid.UUID]entities.MessageState),
	}
}

func (s *State) Set(userId uuid.UUID, newState entities.MessageState) {
	// Lock before setting since this map is shared among all users
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state[userId] = newState
}

func (s *State) Get(userId uuid.UUID) entities.MessageState {
	return s.state[userId]
}

func New(
	auctionService *auction_service.AuctionService,
	userService *user_service.UserService,
	playerService *player_service.PlayerService,
	leagueService *league_service.LeagueService,
) *MessageService {
	state := NewState()

	return &MessageService{
		auctionService,
		userService,
		playerService,
		leagueService,
		state,
	}
}

func (m *MessageService) SendAction(context echo.Context, action entities.Action, userId uuid.UUID, event interface{}) error {
	switch m.state.Get(userId) {
	case entities.STATE_AUCTION_OPENED:
		break
	case entities.STATE_BIDDING:
		m.handleBiddingState(context, action, userId, event)
		break
	case entities.STATE_BIDDING_FINISHED:

		break
	case entities.STATE_INVALID:
	default:

		break
	}

	return nil
}

func (m *MessageService) handleBiddingState(context echo.Context, action entities.Action, userId uuid.UUID, event interface{}) error {
	switch action {
	case entities.ACTION_SEND_MESSAGE:
		message := event.(messenger_entities.WebhookMessage)

		// If the user is sending a bid, parse it for the bid value
		if strings.HasPrefix(message.Text, "bid") {

		}

		break
	case entities.ACTION_SEND_POSTBACK:

		break
	case entities.ACTION_SEND_READ:

		break
	case entities.ACTION_INVALID:
	default:

		break
	}

	return nil
}

func (m *MessageService) CreateWinningBidForPlayerEvent(context echo.Context, winningBid entities.AuctionBid) (messenger_entities.SendEvent, error) {
	player, err := m.playerService.GetPlayerByPlayerId(context, winningBid.PlayerId)
	if err != nil {
		return messenger_entities.SendEvent{}, err
	}

	return messenger_entities.SendEvent{
		Message: messenger_entities.SendMessage{
			Attachment: messenger_entities.Template{
				Type: "template",
				Payload: messenger_entities.TemplatePayload{
					TemplateType: "generic",
					Elements: []messenger_entities.TemplateElements{
						{
							Title:    fmt.Sprintf("%v $%v!", constants.WINNING_BID_TITLE, winningBid.Bid),
							ImageUrl: player.Image,
							Subtitle: fmt.Sprintf("%v | %v | %v \n%v", player.Name, player.Team, player.Position, constants.CLAIM_INSTRUCTIONS),
						},
					},
				},
			},
		},
	}, nil
}

func (m *MessageService) CreateBidsForAuction(context echo.Context, auctionId uuid.UUID) ([]messenger_entities.SendEvent, error) {
	auction, err := m.auctionService.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return nil, err
	}

	// Collect all the users in the auction
	userIds, err := m.leagueService.GetMembersInLeague(context, auction.LeagueId)
	if err != nil {
		return nil, err
	}

	// Get senderPsIds from the userIds
	senderPsIds := make([]string, len(userIds))
	for index, userId := range userIds {
		senderPsId, err := m.userService.GetSenderPsIdFromUserId(context, userId)
		if err != nil {
			return nil, err
		}

		senderPsIds[index] = senderPsId
	}

	// Get all playerIds from the auction's player set
	// TODO: create player set endpoint
	playerIds := []string{
		"44-julio-rodriguez",
		"7-jarred-kelenic",
		"13-bobby-witt",
	}

	// Create player bid template item for each player
	playerBidTemplateElementsMap, err := m.CreatePlayerBidTemplateElementsMap(context, playerIds, senderPsIds, auctionId)
	if err != nil {
		return nil, err
	}

	return m.CreatePlayerBidEvents(context, senderPsIds, playerBidTemplateElementsMap)
}

func (m *MessageService) CreatePlayerBidEvents(context echo.Context, senderPsIds []string, playerBidTemplateElementsMap map[string][]messenger_entities.TemplateElements) ([]messenger_entities.SendEvent, error) {
	events := make([]messenger_entities.SendEvent, len(senderPsIds))
	for index, senderPsId := range senderPsIds {
		sendEvent := messenger_entities.SendEvent{
			Recipient: messenger_entities.Id{
				Id: senderPsId,
			},
			Message: messenger_entities.SendMessage{
				Attachment: messenger_entities.Template{
					Type: "template",
					Payload: messenger_entities.TemplatePayload{
						TemplateType: "generic",
						Elements:     playerBidTemplateElementsMap[senderPsId],
					},
				},
			},
		}

		events[index] = sendEvent
	}

	return events, nil
}

func (m *MessageService) CreatePlayerBidTemplateElementsMap(context echo.Context, playerIds []string, senderPsIds []string, auctionId uuid.UUID) (map[string][]messenger_entities.TemplateElements, error) {
	senderPsIdsTemplateElementMap := make(map[string][]messenger_entities.TemplateElements)

	// For performance reasons, keep a map of the players we already retrieved
	playerMap := make(map[string]*entities.Player)

	for _, senderPsId := range senderPsIds {
		templateElements := make([]messenger_entities.TemplateElements, len(playerIds))

		for index, playerId := range playerIds {
			// If the player was already retrieved, grab from the map
			var player *entities.Player
			if playerMap[playerId] != nil {
				player = playerMap[playerId]
			} else {
				result, err := m.playerService.GetPlayerByPlayerId(context, playerId)
				if err != nil {
					return nil, err
				}
				player = &result
			}

			params := url.Values{}
			params.Add("auction_id", auctionId.String())
			params.Add("player_id", playerId)
			params.Add("sender_ps_id", senderPsId)

			templateElement := messenger_entities.TemplateElements{
				Title:    player.Name,
				ImageUrl: player.Image,
				Subtitle: fmt.Sprintf("%v | %v", player.Team, player.Position),
				Buttons: []messenger_entities.TemplateDefaultAction{
					{
						Type:               "web_url",
						Url:                "https://3638-50-35-81-67.ngrok.io/webview/bid/?" + params.Encode(),
						WebviewHeightRatio: "compact",
						Title:              "Place bid",
					},
					{
						Type:    "postback",
						Payload: "payloadTest",
						Title:   "Testing postback button",
					},
				},
			}

			templateElements[index] = templateElement
		}

		senderPsIdsTemplateElementMap[senderPsId] = templateElements
	}

	return senderPsIdsTemplateElementMap, nil
}
