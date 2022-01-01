package auction_service

import (
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	redis_client "github.com/wilbertthelam/prop-ock/db"
	"github.com/wilbertthelam/prop-ock/entities"
	auction_repo "github.com/wilbertthelam/prop-ock/repos/auction"
	league_service "github.com/wilbertthelam/prop-ock/services/league"
	player_service "github.com/wilbertthelam/prop-ock/services/player"
	user_service "github.com/wilbertthelam/prop-ock/services/user"
	"github.com/wilbertthelam/prop-ock/utils"
)

type AuctionService struct {
	auctionRepo   *auction_repo.AuctionRepo
	userService   *user_service.UserService
	playerService *player_service.PlayerService
	leagueService *league_service.LeagueService
	redisClient   *redis.Client
}

func New(
	auctionRepo *auction_repo.AuctionRepo,
	userService *user_service.UserService,
	playerService *player_service.PlayerService,
	leagueService *league_service.LeagueService,
	redisClient *redis.Client,
) *AuctionService {
	return &AuctionService{
		auctionRepo,
		userService,
		playerService,
		leagueService,
		redisClient,
	}
}

func (a *AuctionService) GetAuctionByAuctionId(context echo.Context, auctionId uuid.UUID) (entities.Auction, error) {
	return a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
}

func (a *AuctionService) GetCurrentAuctionIdByLeagueId(context echo.Context, leagueId uuid.UUID) (uuid.UUID, error) {
	return a.auctionRepo.GetCurrentAuctionIdByLeagueId(context, leagueId)
}

func (a *AuctionService) CreateAuction(context echo.Context, auctionId uuid.UUID, leagueId uuid.UUID, startTime int64, endTime int64) (entities.Auction, error) {
	// Check if there's already an existing auction running for this league
	existingAuctionId, err := a.auctionRepo.GetCurrentAuctionIdByLeagueId(context, leagueId)
	if err != nil {
		return entities.Auction{}, err
	}

	if existingAuctionId != uuid.Nil {
		// Check the auction status, if it is not finished yet then we don't want to start a new one
		existingAuction, err := a.GetAuctionByAuctionId(context, existingAuctionId)
		if err != nil {
			return entities.Auction{}, err
		}

		// Only allow us to create auctions when one is currently not running
		if existingAuction.Status != entities.AUCTION_STATUS_INVALID &&
			existingAuction.Status != entities.AUCTION_STATUS_CLOSED {
			return entities.Auction{}, utils.NewError(utils.ErrorParams{
				Code:    http.StatusBadRequest,
				Message: "an auction is currently running for this league",
				Args: []interface{}{
					"auctionId", existingAuction.Id.String(),
					"leagueId", leagueId.String(),
					"auctionStatus", fmt.Sprintf("%v", existingAuction.Status),
				},
				Err: nil,
			})
		}
	}

	// Create new auction UUID if not provided
	if auctionId == uuid.Nil {
		auctionId = uuid.New()
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

	// Start Redis transaction here to create auction
	err = redis_client.StartTransaction(
		context,
		a.redisClient,
		func() error {
			err = a.auctionRepo.SetLeagueToAuctionRelationship(context, leagueId, auctionId)
			if err != nil {
				return err
			}

			err = a.auctionRepo.CreateAuction(context, auctionId, auction)
			if err != nil {
				return err
			}

			return nil
		},
	)
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
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "cannot start an auction that's not in created state",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"auctionStatus", fmt.Sprintf("%v", auction.Status),
			},
			Err: nil,
		})
	}

	return a.auctionRepo.StartAuction(context, auctionId)
}

// Stop auction
func (a *AuctionService) StopAuction(context echo.Context, auctionId uuid.UUID) error {
	// Check if the auction is already created
	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// Auction can only be stopped if it is in the ACTIVE status
	if auction.Status != entities.AUCTION_STATUS_ACTIVE {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "cannot stop an auction that's not in active state",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"auctionStatus", fmt.Sprintf("%v", auction.Status),
			},
			Err: nil,
		})
	}

	return a.auctionRepo.StopAuction(context, auctionId)
}

// Close auction
func (a *AuctionService) CloseAuction(context echo.Context, auctionId uuid.UUID) error {
	// Check if the auction is already created
	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// Auction can only be closed if it is in the STOPPED status
	if auction.Status != entities.AUCTION_STATUS_STOPPED {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "cannot archive an auction that's not in stopped state",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"auctionStatus", fmt.Sprintf("%v", auction.Status),
			},
			Err: nil,
		})
	}

	return a.auctionRepo.CloseAuction(context, auctionId)
}

// MakeBid sends in a bid for a player by a given user for a specific auction
func (a *AuctionService) MakeBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string, bid int64) error {
	// Make sure bid is positive
	if bid < 0 {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "cannot make a bid with negative value",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
				"playerId", playerId,
				"bid", fmt.Sprintf("%v", bid),
			},
			Err: nil,
		})
	}

	// Check if auction is open and is active
	isAuctionOpen, err := a.ValidateAuctionIsActive(context, auctionId)
	if err != nil {
		return err
	}

	if !isAuctionOpen {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "cannot make bid on a non-active auction",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
				"playerId", playerId,
				"bid", fmt.Sprintf("%v", bid),
			},
			Err: nil,
		})
	}

	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// TODO: Make sure the player exists

	// TODO: Make sure the user exists

	// Make sure the user hasn't already made a bid
	existingBid, err := a.GetBid(context, auctionId, userId, playerId)
	if err != nil {
		return err
	}

	if existingBid >= 0 {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "cannot make another bids on the same player if bid already exists",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
				"playerId", playerId,
				"bid", fmt.Sprintf("%v", bid),
			},
			Err: nil,
		})
	}

	// Place bid updates in transaction as there are multiple updates
	// to multiple keys (for the wallet and for the bid item)
	return redis_client.StartTransaction(
		context,
		a.redisClient,
		func() error {
			// Try removing funds from the user wallet (validation happens inside user service)
			_, err = a.userService.RemoveFundsFromUserWallet(context, userId, auction.LeagueId, bid)
			if err != nil {
				return err
			}

			// Create a bid for the player
			return a.auctionRepo.MakeBid(context, auctionId, userId, playerId, bid)
		},
	)
}

func (a *AuctionService) GetBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string) (int64, error) {
	return a.auctionRepo.GetBid(context, auctionId, userId, playerId)
}

func (a *AuctionService) CancelBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string) error {
	// Check if auction is open and is active
	isAuctionOpen, err := a.ValidateAuctionIsActive(context, auctionId)
	if err != nil {
		return err
	}

	if !isAuctionOpen {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "cannot cancel bid on a non-active auction",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
				"playerId", playerId,
			},
			Err: nil,
		})
	}

	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// Get the prior bid amount
	bid, err := a.GetBid(context, auctionId, userId, playerId)
	if err != nil {
		return err
	}

	// Check to make sure the bid exists before canceling
	if bid < 0 {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "cannot cancel bid that doesn't exist for player",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
				"playerId", playerId,
			},
			Err: nil,
		})
	}

	_, err = a.userService.AddFundsToUserWallet(context, userId, auction.LeagueId, bid)
	if err != nil {
		return err
	}

	return a.auctionRepo.CancelBid(context, auctionId, userId, playerId)
}

func (a *AuctionService) ValidateAuctionIsActive(context echo.Context, auctionId uuid.UUID) (bool, error) {
	// Check if auction is open and is active
	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return false, err
	}

	return auction.Status == entities.AUCTION_STATUS_ACTIVE, nil
}

func (a *AuctionService) GetAllUserBids(context echo.Context, auctionId uuid.UUID, userId uuid.UUID) (map[string]int64, error) {
	return a.auctionRepo.GetAllUserBids(context, auctionId, userId)
}

func (a *AuctionService) GetAuctionResults(context echo.Context, auctionId uuid.UUID) (map[string][]entities.AuctionBid, error) {
	_, err := a.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return nil, err
	}

	return a.auctionRepo.GetAuctionResults(context, auctionId)
}

func (a *AuctionService) ProcessAuction(context echo.Context, auctionId uuid.UUID) error {
	// Make sure auction is stopped first
	// Check if the auction is created
	auction, err := a.auctionRepo.GetAuctionByAuctionId(context, auctionId)
	if err != nil {
		return err
	}

	// Auction can only be processed if it is in the STOPPED status
	if auction.Status != entities.AUCTION_STATUS_STOPPED {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusBadRequest,
			Message: "cannot close an auction that is not stopped",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"auctionStatus", fmt.Sprintf("%v", auction.Status),
			},
			Err: nil,
		})
	}

	leagueId := auction.LeagueId

	// Get users inside this auction
	userIds, err := a.leagueService.GetMembersInLeague(context, leagueId)
	if err != nil {
		return nil
	}

	// Create a map keyed on playerId with a value of a list of the highest
	// bids ({ userId, bid })
	playerWinningBidsMap := make(map[string][]entities.AuctionBid)
	playerLosingBidsMap := make(map[string][]entities.AuctionBid)

	for _, userId := range userIds {
		bids, err := a.GetAllUserBids(context, auctionId, userId)
		if err != nil {
			return err
		}

		for playerId, bid := range bids {
			auctionBid := entities.AuctionBid{
				UserId:    userId,
				Bid:       bid,
				AuctionId: auctionId,
				PlayerId:  playerId,
			}

			// If no value instantiated yet, then they are the highest bid
			highBidList, ok := playerWinningBidsMap[playerId]
			if !ok {
				playerWinningBidsMap[playerId] = []entities.AuctionBid{auctionBid}
				continue
			}

			// The auctionBid list holds all bids which have the highest value (including ties)
			// If there are more than 1 bid (ties), we'll only have to compare against the first bid
			currentHighestBid := highBidList[0]

			// If the current highest bids are less than the next bid,
			// then move the current highest bid into the losing bids map
			// and we replace the current highest bid with this next bid
			if currentHighestBid.Bid < bid {
				playerLosingBids, ok := playerLosingBidsMap[playerId]
				if !ok {
					playerLosingBidsMap[playerId] = highBidList
				} else {
					playerLosingBidsMap[playerId] = append(playerLosingBids, highBidList...)
				}

				playerWinningBidsMap[playerId] = []entities.AuctionBid{auctionBid}
				continue
			}

			// If the bids are tied, we just add it into the list of highest bids
			if currentHighestBid.Bid == bid {
				playerWinningBidsMap[playerId] = append(playerWinningBidsMap[playerId], auctionBid)
				continue
			}

			// If the bid is lower than the highest bids, just add it to the losing bids list
			if currentHighestBid.Bid > bid {
				playerLosingBids, ok := playerLosingBidsMap[playerId]
				if !ok {
					playerLosingBidsMap[playerId] = []entities.AuctionBid{auctionBid}
				} else {
					playerLosingBidsMap[playerId] = append(playerLosingBids, auctionBid)
				}
			}

		}
	}

	// Save the auction results for retrieval
	err = a.auctionRepo.SaveAuctionResult(context, auctionId, playerWinningBidsMap)
	if err != nil {
		return err
	}

	// Issue refunds for failed bids
	userTotalRefundAmount := make(map[uuid.UUID]int64)
	for _, playerBids := range playerLosingBidsMap {
		for _, playerBid := range playerBids {
			userId := playerBid.UserId
			bidAmount := playerBid.Bid

			totalRefundAmount, ok := userTotalRefundAmount[userId]
			if !ok {
				userTotalRefundAmount[userId] = bidAmount
			} else {
				userTotalRefundAmount[userId] = totalRefundAmount + bidAmount
			}
		}
	}

	for userId, bid := range userTotalRefundAmount {
		updatedFunds, err := a.userService.AddFundsToUserWallet(context, userId, leagueId, bid)
		if err != nil {
			return err
		}

		context.Logger().Infof("updated funds after bid returns: userId: %v, funds: %v", userId, updatedFunds)
	}

	// Close auction once it's been processed
	err = a.CloseAuction(context, auctionId)
	if err != nil {
		return err
	}

	return nil
}
