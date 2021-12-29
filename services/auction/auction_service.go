package auction_service

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
	messenger_entities "github.com/wilbertthelam/prop-ock/entities/messenger"
	auction_repo "github.com/wilbertthelam/prop-ock/repos/auction"
	league_service "github.com/wilbertthelam/prop-ock/services/league"
	player_service "github.com/wilbertthelam/prop-ock/services/player"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
)

type AuctionService struct {
	auctionRepo   *auction_repo.AuctionRepo
	userService   *user_service.UserService
	playerService *player_service.PlayerService
	leagueService *league_service.LeagueService
}

func New(
	auctionRepo *auction_repo.AuctionRepo,
	userService *user_service.UserService,
	playerService *player_service.PlayerService,
	leagueService *league_service.LeagueService,
) *AuctionService {
	return &AuctionService{
		auctionRepo,
		userService,
		playerService,
		leagueService,
	}
}

func GetName() string {
	return "auction_service"
}

func (a *AuctionService) PlaceBid(context echo.Context, leagueId uuid.UUID, playerId uuid.UUID, userId uuid.UUID, bid int) error {
	// Check to see if auction is still going

	// Check if user is in league

	// Check if user has enough value to spend on this bid

	// Check if player is valid

	// Check if bid amount is valid
	if bid < 0 {
		return fmt.Errorf("invalid bid less than 0: %d", bid)
	}

	// Send bid into Redis

	return nil
}

func (a *AuctionService) GetAuctionByAuctionId(context echo.Context, auctionId uuid.UUID) (entities.Auction, error) {
	return a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
}

func (a *AuctionService) GetAuctionIdByLeagueId(context echo.Context, leagueId uuid.UUID) (uuid.UUID, error) {
	return a.auctionRepo.GetAuctionIdByLeagueId(context, leagueId)
}

func (a *AuctionService) CreateAuction(context echo.Context, auctionId uuid.UUID, leagueId uuid.UUID, startTime int64, endTime int64) (entities.Auction, error) {
	// Check if there's already an existing auction running for this league
	existingAuctionId, err := a.auctionRepo.GetAuctionIdByLeagueId(context, leagueId)
	if err != nil {
		return entities.Auction{}, err
	}

	if existingAuctionId != uuid.Nil {
		// Check the auction status, if it is not finished yet then we don't want to start a new one
		existingAuction, err := a.GetAuctionByAuctionId(context, existingAuctionId)
		if err != nil {
			return entities.Auction{}, err
		}

		// Only allow us to
		if existingAuction.Status != entities.AUCTION_STATUS_INVALID &&
			existingAuction.Status != entities.AUCTION_STATUS_ARCHIVED {
			return entities.Auction{}, fmt.Errorf("an auction is currently running for this league: %v, auctionId: %v, auctionStatus: %v", leagueId, auctionId, existingAuction.Status)
		}
	}

	// Create new auction UUID if not provided
	if auctionId == uuid.Nil {
		auctionId = uuid.New()
	}

	// Start Redis transaction here to create auction

	err = a.auctionRepo.SetLeagueToAuctionRelationship(context, leagueId, auctionId)
	if err != nil {
		return entities.Auction{}, err
	}

	// Create the auction object
	auction := entities.Auction{
		Id:        auctionId,
		LeagueId:  leagueId,
		StartTime: startTime,
		EndTime:   endTime,
		Name:      "",
		Status:    entities.AUCTION_STATUS_CREATED,
	}

	err = a.auctionRepo.CreateAuction(context, auctionId, auction)
	if err != nil {
		return entities.Auction{}, err
	}

	return auction, nil
}

// Start auction
func (a *AuctionService) StartAuction(context echo.Context, auctionId uuid.UUID) error {
	// Check if the auction is already created
	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// Auction can only be started if it is in the CREATED status
	if auction.Status != entities.AUCTION_STATUS_CREATED {
		return fmt.Errorf("cannot start an auction that's not in created state: %v", auction.Status)
	}

	return a.auctionRepo.StartAuction(context, auctionId)
}

// Close auction
func (a *AuctionService) CloseAuction(context echo.Context, auctionId uuid.UUID) error {
	// Check if the auction is already created
	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// Auction can only be closed if it is in the ACTIVE status
	if auction.Status != entities.AUCTION_STATUS_ACTIVE {
		return fmt.Errorf("cannot close an auction that's not in active state: %v", auction.Status)
	}

	return a.auctionRepo.CloseAuction(context, auctionId)
}

// Archive auction
func (a *AuctionService) ArchiveAuction(context echo.Context, auctionId uuid.UUID) error {
	// Check if the auction is already created
	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// Auction can only be closed if it is in the CLOSED status
	if auction.Status != entities.AUCTION_STATUS_CLOSED {
		return fmt.Errorf("cannot archive an auction that's not in closed state: %v", auction.Status)
	}

	return a.auctionRepo.ArchiveAuction(context, auctionId)
}

// MakeBid sends in a bid for a player by a given user for a specific auction
func (a *AuctionService) MakeBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string, bid int64) error {
	// Make sure bid is positive
	if bid < 0 {
		return fmt.Errorf("bid cannot be negative: %v", bid)
	}

	// Check if auction is open and is active
	isAuctionOpen, err := a.ValidateAuctionIsOpen(context, auctionId)
	if err != nil {
		return err
	}

	if !isAuctionOpen {
		return fmt.Errorf("auction is not currently open: %v", auctionId)
	}

	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// TODO: Make sure the player exists

	// TODO: Make sure the user exists

	// Try removing funds from the user wallet (validation happens inside user service)
	_, err = a.userService.RemoveFundsFromUserWallet(context, userId, auction.LeagueId, bid)
	if err != nil {
		return err
	}

	// Create a bid for the player
	return a.auctionRepo.MakeBid(context, auctionId, userId, playerId, bid)
}

func (a *AuctionService) GetBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string) (int64, error) {
	return a.auctionRepo.GetBid(context, auctionId, userId, playerId)
}

func (a *AuctionService) CancelBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string) error {
	// Check if auction is open and is active
	isAuctionOpen, err := a.ValidateAuctionIsOpen(context, auctionId)
	if err != nil {
		return err
	}

	if !isAuctionOpen {
		return fmt.Errorf("auction is not currently open: %v", auctionId)
	}

	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// Get the prior bid amount
	bid, err := a.auctionRepo.GetBid(context, auctionId, userId, playerId)
	if err != nil {
		return err
	}

	_, err = a.userService.AddFundsToUserWallet(context, userId, auction.LeagueId, bid)
	if err != nil {
		return err
	}

	return a.auctionRepo.CancelBid(context, auctionId, userId, playerId)
}

func (a *AuctionService) ValidateAuctionIsOpen(context echo.Context, auctionId uuid.UUID) (bool, error) {
	// Check if auction is open and is active
	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return false, err
	}

	return auction.Status == entities.AUCTION_STATUS_ACTIVE, nil
}

func (a *AuctionService) CreateBidsForAuction(context echo.Context, auctionId uuid.UUID) ([]messenger_entities.SendEvent, error) {
	auction, err := a.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return nil, err
	}

	// Collect all the users in the auction
	userIds, err := a.leagueService.GetMembersInLeague(context, auction.LeagueId)
	if err != nil {
		return nil, err
	}

	// Get senderPsIds from the userIds
	senderPsIds := make([]string, len(userIds))
	for index, userId := range userIds {
		senderPsId, err := a.userService.GetSenderPsIdFromUserId(context, userId)
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
	playerBidTemplateElementsMap, err := a.CreatePlayerBidTemplateElementsMap(context, playerIds, senderPsIds, auctionId)
	if err != nil {
		return nil, err
	}

	return a.CreatePlayerBidEvents(context, senderPsIds, playerBidTemplateElementsMap)
}

func (a *AuctionService) CreatePlayerBidEvents(context echo.Context, senderPsIds []string, playerBidTemplateElementsMap map[string][]messenger_entities.TemplateElements) ([]messenger_entities.SendEvent, error) {
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

func (a *AuctionService) CreatePlayerBidTemplateElementsMap(context echo.Context, playerIds []string, senderPsIds []string, auctionId uuid.UUID) (map[string][]messenger_entities.TemplateElements, error) {
	senderPsIdsTemplateElementMap := make(map[string][]messenger_entities.TemplateElements, len(senderPsIds))

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
				result, err := a.playerService.GetPlayerByPlayerId(context, playerId)
				if err != nil {
					return nil, err
				}
				player = &result
			}

			params := url.Values{}
			params.Add("auctionId", auctionId.String())
			params.Add("playerId", playerId)
			params.Add("senderPsId", senderPsId)

			context.Logger().Info("https://5955-50-35-81-67.ngrok.io/webview/bid/?" + params.Encode())

			templateElement := messenger_entities.TemplateElements{
				Title:    player.Name,
				ImageUrl: player.Image,
				Subtitle: fmt.Sprintf("%v | %v", player.Team, player.Position),
				Buttons: []messenger_entities.TemplateDefaultAction{
					{
						Type:               "web_url",
						Url:                "https://5955-50-35-81-67.ngrok.io/webview/bid/?" + params.Encode(),
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
