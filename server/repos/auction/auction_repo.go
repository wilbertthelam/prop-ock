package auction_repo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	redis_client "github.com/wilbertthelam/prop-ock/db"
	"github.com/wilbertthelam/prop-ock/entities"
	"github.com/wilbertthelam/prop-ock/utils"
)

type AuctionRepo struct {
	redisClient *redis.Client
}

func New(redisClient *redis.Client) *AuctionRepo {
	return &AuctionRepo{
		redisClient,
	}
}

func generateAuctionRedisKey(auctionId uuid.UUID) string {
	return fmt.Sprintf("auction:auction_id:%v", auctionId.String())
}

func generateLeagueToActiveAuctionRelationshipRedisKey(leagueId uuid.UUID) string {
	return fmt.Sprintf("relationship:league_to_current_auction:league_id:%v", leagueId.String())
}

func generateBidRedisKey(auctionId uuid.UUID, userId uuid.UUID) string {
	return fmt.Sprintf("bid:auction_id:%v:user_id:%v", auctionId.String(), userId.String())
}

func generateAuctionResultsRedisKey(auctionId uuid.UUID) string {
	return fmt.Sprintf("result:auction_id:%v", auctionId.String())
}

func (a *AuctionRepo) GetAuctionByAuctionId(context echo.Context, auctionId uuid.UUID) (entities.Auction, error) {
	// Query Redis for the auction
	redisAuction, err := a.redisClient.HGetAll(
		context.Request().Context(),
		generateAuctionRedisKey(auctionId),
	).Result()
	if err != nil {
		return entities.Auction{}, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to get auction",
			Args: []interface{}{
				"auctionId", auctionId.String(),
			},
			Err: err,
		})
	}

	if len(redisAuction) == 0 {
		return entities.Auction{}, utils.NewError(utils.ErrorParams{
			Code:    http.StatusNotFound,
			Message: "no auction found",
			Args: []interface{}{
				"auctionId", auctionId.String(),
			},
			Err: nil,
		})
	}

	startTime, err := strconv.ParseInt(redisAuction["start_time"], 10, 64)
	if err != nil {
		return entities.Auction{}, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to parse start time for auction",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"startTime", redisAuction["start_time"],
			},
			Err: err,
		})
	}

	endTime, err := strconv.ParseInt(redisAuction["end_time"], 10, 64)
	if err != nil {
		return entities.Auction{}, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to parse end time for auction",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"endTime", redisAuction["end_time"],
			},
			Err: err,
		})
	}

	status, err := strconv.ParseInt(redisAuction["status"], 10, 64)
	if err != nil {
		return entities.Auction{}, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to parse status for auction",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"status", redisAuction["status"],
			},
			Err: err,
		})
	}

	auction := entities.Auction{
		Id:        uuid.Must(uuid.Parse(redisAuction["id"])),
		LeagueId:  uuid.Must(uuid.Parse(redisAuction["league_id"])),
		StartTime: startTime,
		EndTime:   endTime,
		Status:    entities.AuctionStatus(status),
		Name:      redisAuction["name"],
		Notes:     redisAuction["notes"],
	}

	return auction, nil
}

func (a *AuctionRepo) GetCurrentAuctionIdByLeagueId(context echo.Context, leagueId uuid.UUID) (uuid.UUID, error) {
	// Query Redis for the leagueId
	auctionId, err := a.redisClient.Get(
		context.Request().Context(),
		generateLeagueToActiveAuctionRelationshipRedisKey(leagueId),
	).Result()

	// If Redis key doesn't exist, then auction doesn't exist
	if err == redis.Nil {
		return uuid.Nil, utils.NewError(utils.ErrorParams{
			Code:    http.StatusNotFound,
			Message: "current auction does not exist for league id",
			Args: []interface{}{
				"leagueId", leagueId.String(),
			},
			Err: nil,
		})
	}

	if err != nil {
		return uuid.Nil, utils.NewError(utils.ErrorParams{
			Code:    http.StatusNotFound,
			Message: "failed to get league to current auction relationship",
			Args: []interface{}{
				"leagueId", leagueId.String(),
				"auctionId", auctionId,
			},
			Err: err,
		})
	}

	if auctionId == "" {
		return uuid.Nil, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "dangling league to current auction relationship",
			Args: []interface{}{
				"leagueId", leagueId.String(),
			},
			Err: nil,
		})
	}

	return uuid.MustParse(auctionId), nil
}

func (a *AuctionRepo) SetLeagueToAuctionRelationship(context echo.Context, leagueId uuid.UUID, auctionId uuid.UUID) error {
	// Upsert the league to auction relationship
	_, err := redis_client.
		GetCmdable(context, a.redisClient).
		Set(
			context.Request().Context(),
			generateLeagueToActiveAuctionRelationshipRedisKey(leagueId),
			auctionId.String(),
			0,
		).Result()
	if err != nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to set league to auction relationship",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"leagueId", leagueId.String(),
			},
			Err: err,
		})
	}

	return nil
}

func (a *AuctionRepo) CreateAuction(context echo.Context, auctionId uuid.UUID, auction entities.Auction) error {
	// Format for database inserting (array of strings where even index is key, odd index is value)
	redisAuctionKeyValuePairs := []string{
		"id", auction.Id.String(),
		"league_id", auction.LeagueId.String(),
		"start_time", strconv.FormatInt(auction.StartTime, 10),
		"end_time", strconv.FormatInt(auction.EndTime, 10),
		"status", strconv.FormatInt(int64(auction.Status), 10),
		"name", auction.Name,
	}

	err := a.updateAuction(context, auctionId, redisAuctionKeyValuePairs)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuctionRepo) StartAuction(context echo.Context, auctionId uuid.UUID) error {
	redisStatusKeyValuePair := []string{
		"status", strconv.FormatInt(int64(entities.AUCTION_STATUS_ACTIVE), 10),
	}

	err := a.updateAuction(context, auctionId, redisStatusKeyValuePair)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuctionRepo) StopAuction(context echo.Context, auctionId uuid.UUID) error {
	redisStatusKeyValuePair := []string{
		"status", strconv.FormatInt(int64(entities.AUCTION_STATUS_STOPPED), 10),
	}

	err := a.updateAuction(context, auctionId, redisStatusKeyValuePair)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuctionRepo) CloseAuction(context echo.Context, auctionId uuid.UUID) error {
	redisStatusKeyValuePair := []string{
		"status", strconv.FormatInt(int64(entities.AUCTION_STATUS_CLOSED), 10),
	}

	err := a.updateAuction(context, auctionId, redisStatusKeyValuePair)
	if err != nil {
		return err
	}

	return nil
}

func (a *AuctionRepo) GetAllUserBids(context echo.Context, auctionId uuid.UUID, userId uuid.UUID) (map[string]int64, error) {
	rawPlayerBids, err := a.redisClient.HGetAll(
		context.Request().Context(),
		generateBidRedisKey(auctionId, userId),
	).Result()
	if err != nil {
		return nil, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to get all of a user's bids",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
			},
			Err: err,
		})
	}

	bids := make(map[string]int64)
	for playerId, bidString := range rawPlayerBids {
		bid, err := strconv.ParseInt(bidString, 10, 64)
		if err != nil {
			return nil, utils.NewError(utils.ErrorParams{
				Code:    http.StatusInternalServerError,
				Message: "failed to parse a bid when trying to get all of a user's bids",
				Args: []interface{}{
					"auctionId", auctionId.String(),
					"userId", userId.String(),
				},
				Err: err,
			})
		}

		bids[playerId] = bid
	}

	return bids, nil
}

func (a *AuctionRepo) GetBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string) (int64, error) {
	bidString, err := a.redisClient.HGet(
		context.Request().Context(),
		generateBidRedisKey(auctionId, userId),
		playerId,
	).Result()

	// If Redis key doesn't exist, then bid doesn't exist
	if err == redis.Nil {
		return -1, nil
	}

	if err != nil {
		return -1, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to get a bid for a player",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
				"playerId", playerId,
			},
			Err: err,
		})
	}

	bid, err := strconv.ParseInt(bidString, 10, 64)
	if err != nil {
		return -1, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to parse bid amount for a player",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
				"playerId", playerId,
				"bid", bidString,
			},
			Err: err,
		})
	}

	return bid, nil
}

func (a *AuctionRepo) MakeBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string, bid int64) error {
	redisBidKeyValuePair := []string{
		playerId, strconv.FormatInt(bid, 10),
	}

	_, err := redis_client.
		GetCmdable(context, a.redisClient).
		HSet(
			context.Request().Context(),
			generateBidRedisKey(auctionId, userId),
			redisBidKeyValuePair,
		).Result()
	if err != nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to make a bid",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
				"playerId", playerId,
				"bid", fmt.Sprintf("%v", bid),
			},
			Err: err,
		})
	}

	return nil
}

func (a *AuctionRepo) CancelBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string) error {
	_, err := redis_client.
		GetCmdable(context, a.redisClient).
		HDel(
			context.Request().Context(),
			generateBidRedisKey(auctionId, userId),
			playerId,
		).Result()
	if err != nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to cancel a bid",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"userId", userId.String(),
				"playerId", playerId,
			},
			Err: err,
		})
	}

	return nil
}

func (a *AuctionRepo) updateAuction(context echo.Context, auctionId uuid.UUID, keyValuePairs []string) error {
	_, err := redis_client.
		GetCmdable(context, a.redisClient).
		HSet(
			context.Request().Context(),
			generateAuctionRedisKey(auctionId),
			keyValuePairs,
		).Result()
	if err != nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to update auction fields",
			Args: append(
				[]interface{}{"auctionId", auctionId.String()},
				utils.MapStringSliceToInterfaceSlice(keyValuePairs)...,
			),
			Err: err,
		})
	}

	return nil
}

// SaveAuctionResults stores the processed result of the auction into the DB
func (a *AuctionRepo) SaveAuctionResult(context echo.Context, auctionId uuid.UUID, playerBidMap map[string][]entities.AuctionBid) error {
	playerBidMapSize := len(playerBidMap)
	if playerBidMapSize == 0 {
		return nil
	}

	serializedPlayerBidMap := make(map[string]string, playerBidMapSize)
	for playerId, bid := range playerBidMap {
		serializedBid, err := json.Marshal(bid)
		if err != nil {
			return utils.NewError(utils.ErrorParams{
				Code:    http.StatusInternalServerError,
				Message: "failed to save marshal player bid in saving auction result",
				Args: []interface{}{
					"auctionId", auctionId.String(),
					"playerId", playerId,
				},
				Err: err,
			})
		}
		serializedPlayerBidMap[playerId] = string(serializedBid)
	}

	_, err := a.redisClient.HSet(
		context.Request().Context(),
		generateAuctionResultsRedisKey(auctionId),
		serializedPlayerBidMap,
	).Result()
	if err != nil {
		return utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to save auction results",
			Args: []interface{}{
				"auctionId", auctionId.String(),
				"playerBidMap", fmt.Sprintf("%+v", serializedPlayerBidMap),
			},
			Err: err,
		})
	}

	return nil
}

func (a *AuctionRepo) GetAuctionResults(context echo.Context, auctionId uuid.UUID) (map[string][]entities.AuctionBid, error) {
	rawResults, err := a.redisClient.HGetAll(
		context.Request().Context(),
		generateAuctionResultsRedisKey(auctionId),
	).Result()
	if err != nil {
		return nil, utils.NewError(utils.ErrorParams{
			Code:    http.StatusInternalServerError,
			Message: "failed to get auction results",
			Args: []interface{}{
				"auctionId", auctionId.String(),
			},
			Err: err,
		})
	}

	auctionResults := make(map[string][]entities.AuctionBid)
	for playerId, serializedBid := range rawResults {
		var auctionBid []entities.AuctionBid
		err := json.Unmarshal([]byte(serializedBid), &auctionBid)
		if err != nil {
			return nil, utils.NewError(utils.ErrorParams{
				Code:    http.StatusInternalServerError,
				Message: "failed to unmarshal auction bid in get auction results",
				Args: []interface{}{
					"auctionId", auctionId.String(),
					"playerId", playerId,
					"serializedBid", serializedBid,
				},
				Err: err,
			})
		}

		auctionResults[playerId] = auctionBid
	}

	return auctionResults, nil
}
