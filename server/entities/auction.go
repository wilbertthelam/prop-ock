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
	// AUCTION_STATUS_STOPPED is when the auction is no longer taking bids
	AUCTION_STATUS_STOPPED AuctionStatus = 3
	// AUCTION_STATUS_CLOSED is when the auction has been processed and can be deleted
	AUCTION_STATUS_CLOSED AuctionStatus = 4
)

type Auction struct {
	Id          uuid.UUID     `json:"auction_id,omitempty"`
	LeagueId    uuid.UUID     `json:"league_id,omitempty"`
	PlayerSetId uuid.UUID     `json:"player_set_id,omitempty"`
	StartTime   int64         `json:"start_time,omitempty"`
	EndTime     int64         `json:"end_time,omitempty"`
	Status      AuctionStatus `json:"status,omitempty"`
	Name        string        `json:"name,omitempty"`
	Notes       string        `json:"notes,omitempty"`
}

type AuctionBid struct {
	Id        uuid.UUID `json:"id,omitempty"`
	AuctionId uuid.UUID `json:"auction_id,omitempty"`
	UserId    uuid.UUID `json:"user_id,omitempty"`
	PlayerId  string    `json:"player_id,omitempty"`
	Bid       int64     `json:"bid,omitempty"`
}
