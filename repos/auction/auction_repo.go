package auction_repo

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/wilbertthelam/prop-ock/entities"
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
	if err != nil || redisAuction == nil {
		return entities.Auction{}, fmt.Errorf("no auction found")
	}

	startTime, err := strconv.ParseInt(redisAuction["start_time"], 10, 64)
	if err != nil {
		return entities.Auction{}, fmt.Errorf("failed to parse start time %v, for auction: %v, err: %+v", redisAuction["start_time"], auctionId, err)
	}

	endTime, err := strconv.ParseInt(redisAuction["end_time"], 10, 64)
	if err != nil {
		return entities.Auction{}, fmt.Errorf("failed to parse end time %v, for auction: %v, err: %+v", redisAuction["end_time"], auctionId, err)
	}

	status, err := strconv.ParseInt(redisAuction["status"], 10, 64)
	if err != nil {
		return entities.Auction{}, fmt.Errorf("failed to parse status %v, for auction: %v, err: %+v", redisAuction["status"], auctionId, err)
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
		return uuid.Nil, nil
	}

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get league to curent auction relationship: error: %+v, auctionId: %v, leagueId: %v", err, auctionId, leagueId)
	}

	if auctionId == "" {
		return uuid.Nil, fmt.Errorf("dangling league to current auction relationship: leagueId: %v", leagueId)
	}

	return uuid.MustParse(auctionId), nil
}

func (a *AuctionRepo) SetLeagueToAuctionRelationship(context echo.Context, leagueId uuid.UUID, auctionId uuid.UUID) error {
	// Upsert the league to auction relationship
	_, err := a.redisClient.Set(
		context.Request().Context(),
		generateLeagueToActiveAuctionRelationshipRedisKey(leagueId),
		auctionId.String(),
		0,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to set create league to auction relationship: error: %+v, auctionId: %v, leagueId: %v", err, auctionId, leagueId)
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
		return nil, fmt.Errorf("failed to get bid: error: %+v, auctionId: %v, userId: %v", err, auctionId, userId)
	}

	bids := make(map[string]int64)
	for playerId, bidString := range rawPlayerBids {
		bid, err := strconv.ParseInt(bidString, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bid in getting all bids %v, for auction: %v, err: %+v", bidString, auctionId, err)
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
		return -1, fmt.Errorf("failed to get bid: error: %+v, auctionId: %v, userId: %v, playerId %v", err, auctionId, userId, playerId)
	}

	bid, err := strconv.ParseInt(bidString, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("failed to parse bid %v, for auction: %v, err: %+v", bidString, auctionId, err)
	}

	return bid, nil
}

func (a *AuctionRepo) MakeBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string, bid int64) error {
	redisBidKeyValuePair := []string{
		playerId, strconv.FormatInt(bid, 10),
	}

	_, err := a.redisClient.HSet(
		context.Request().Context(),
		generateBidRedisKey(auctionId, userId),
		redisBidKeyValuePair,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to set bid: error: %+v, auctionId: %v, userId: %v, playerId %v", err, auctionId, userId, playerId)
	}

	return nil
}

func (a *AuctionRepo) CancelBid(context echo.Context, auctionId uuid.UUID, userId uuid.UUID, playerId string) error {
	_, err := a.redisClient.HDel(
		context.Request().Context(),
		generateBidRedisKey(auctionId, userId),
		playerId,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to cancel bid: error: %+v, auctionId: %v, userId: %v, playerId %v", err, auctionId, userId, playerId)
	}

	return nil
}

func (a *AuctionRepo) updateAuction(context echo.Context, auctionId uuid.UUID, keyValuePairs []string) error {
	_, err := a.redisClient.HSet(
		context.Request().Context(),
		generateAuctionRedisKey(auctionId),
		keyValuePairs,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to update auction fields: error: %+v, auctionId: %v, keyValuePairs: %+v", err, auctionId, keyValuePairs)
	}

	return nil
}

// SaveAuctionResults stores the processed result of the auction into the DB
func (a *AuctionRepo) SaveAuctionResult(context echo.Context, auctionId uuid.UUID, playerBidMap map[string][]entities.AuctionBid) error {
	playerBidMapSize := len(playerBidMap)
	if playerBidMapSize == 0 {
		context.Logger().Info("no bids were placed for any players: auctionId: %v", auctionId)
		return nil
	}

	serializedPlayerBidMap := make(map[string]string, playerBidMapSize)
	for playerId, bid := range playerBidMap {
		serializedBid, err := json.Marshal(bid)
		if err != nil {
			return err
		}
		serializedPlayerBidMap[playerId] = string(serializedBid)
	}

	_, err := a.redisClient.HSet(
		context.Request().Context(),
		generateAuctionResultsRedisKey(auctionId),
		serializedPlayerBidMap,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to save auction results: error: %+v, auctionId: %v, playerBidMap: %+v", err, auctionId, playerBidMap)
	}

	return nil
}
