package auction_service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
	auction_repo "github.com/wilbertthelam/prop-ock/repos/auction"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
)

type AuctionService struct {
	auctionRepo *auction_repo.AuctionRepo
	userService *user_service.UserService
}

func New(auctionRepo *auction_repo.AuctionRepo, userService *user_service.UserService) *AuctionService {
	return &AuctionService{
		auctionRepo,
		userService,
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

func (a *AuctionService) CreateAuction(context echo.Context, auctionId uuid.UUID, leagueId uuid.UUID, startTime int64, endTime int64) (entities.Auction, error) {
	// Check if there's already an existing auction running for this league
	existingAuctionId, err := a.auctionRepo.GetAuctionIdByLeagueId(context, leagueId)
	if err != nil {
		return entities.Auction{}, err
	}

	if existingAuctionId != uuid.Nil {
		return entities.Auction{}, fmt.Errorf("an auction is already active for this league: %v", leagueId)
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

	// Try removing funds from the user wallet
	hasEnoughFunds, err := a.userService.ValidateUserHasEnoughFunds(context, userId, auction.LeagueId, bid)
	if err != nil {
		return err
	}

	if !hasEnoughFunds {
		return fmt.Errorf("wallet does not have enough funds to remove value: %v", bid)
	}

	// Create a bid for the player
	return a.auctionRepo.MakeBid(context, auctionId, userId, playerId, bid)
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

	return nil
}

func (a *AuctionService) ValidateAuctionIsOpen(context echo.Context, auctionId uuid.UUID) (bool, error) {
	// Check if auction is open and is active
	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return false, err
	}

	return auction.Status == entities.AUCTION_STATUS_ACTIVE, nil
}
