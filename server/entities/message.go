package entities

type Action int64

const (
	ACTION_INVALID       Action = 0
	ACTION_SEND_MESSAGE  Action = 1
	ACTION_SEND_POSTBACK Action = 2
	ACTION_SEND_READ     Action = 3
)

type MessageState int64

const (
	STATE_INVALID          MessageState = 0
	STATE_AUCTION_OPENED   MessageState = 1
	STATE_BIDDING          MessageState = 2
	STATE_BIDDING_FINISHED MessageState = 3
)
