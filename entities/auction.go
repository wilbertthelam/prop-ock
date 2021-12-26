package entities

import (
	"github.com/google/uuid"
)

type AuctionStatus int64

const (
	AUCTION_STATUS_INVALID AuctionStatus = 0
	// AUCTION_STATUS_CREATED is when the auction has been created but not started
	AUCTION_STATUS_CREATED AuctionStatus = 1
	// AUCTION_STATUS_ACTIVE is when the auction is in progress and taking bids
	AUCTION_STATUS_ACTIVE AuctionStatus = 2
	// AUCTION_STATUS_CLOSED is when the auction is no longer taking bids
	AUCTION_STATUS_CLOSED AuctionStatus = 3
	// AUCTION_STATUS_COMPLETED is when the auction has been processed and can be deleted
	AUCTION_STATUS_ARCHIVED AuctionStatus = 4
)

type Auction struct {
	Id        uuid.UUID
	LeagueId  uuid.UUID
	StartTime int64
	EndTime   int64
	Status    AuctionStatus
	Name      string
	Notes     string
}
