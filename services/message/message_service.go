package message_service

import (
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
	auction_service "github.com/wilbertthelam/prop-ock/services/auction"
)

type MessageService struct {
	auctionService *auction_service.AuctionService
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

func New(auctionService *auction_service.AuctionService) *MessageService {
	state := NewState()

	return &MessageService{
		auctionService,
		state,
	}
}

func GetName() string {
	return "message_service"
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
		message := event.(entities.MessengerWebhookMessageEvent)

		// If the user is sending a bid, parse it for the bid value
		if strings.HasPrefix(message.Message.Text, "bid") {

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
