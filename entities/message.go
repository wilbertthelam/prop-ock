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

type MessengerWebhookBody struct {
	Object string
	Entry  []MessengerWebhookEntry
}

type MessengerWebhookEntry struct {
	Messaging []MessengerWebhookEvent
}

type MessengerWebhookEvent struct {
	Sender   MessengerWebhookEventSender
	Message  MessengerWebhookMessageEvent  `json:"message,omitempty"`
	Postback MessengerWebhookPostbackEvent `json:"postback,omitempty"`
	Read     MessengerWebhookReadEvent     `json:"read,omitempty"`
}

type MessengerWebhookMessageEvent struct {
	Sender    MessengerId
	Recipient MessengerId
	Timestamp int64
	Message   MessengerWebhookMessage
}

type MessengerWebhookMessage struct {
	Mid  string
	Text string
}

type MessengerId struct {
	Id string
}

type MessengerWebhookPostbackEvent struct{}

type MessengerWebhookReadEvent struct{}

type MessengerWebhookEventSender struct {
	Id string
}
