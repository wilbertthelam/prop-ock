package messenger_entities

type WebhookBody struct {
	Object string         `json:"object,omitempty"`
	Entry  []WebhookEntry `json:"entry,omitempty"`
}

type WebhookEntry struct {
	Messaging []WebhookEvent `json:"messaging,omitempty"`
}

type WebhookEvent struct {
	Sender    Id
	Timestamp int64
	Message   WebhookMessage  `json:"message,omitempty"`
	Postback  WebhookPostback `json:"postback,omitempty"`
	Read      WebhookRead     `json:"read,omitempty"`
}

type WebhookMessage struct {
	Mid  string `json:"mid,omitempty"`
	Text string `json:"text,omitempty"`
}

type WebhookPostback struct {
	Title   string `json:"title,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type WebhookRead struct{}

type WebhookBidPostBody struct {
	PlayerId   string `json:"player_id,omitempty"`
	SenderPsId string `json:"sender_ps_id,omitempty"`
	AuctionId  string `json:"auction_id,omitempty"`
	Bid        int64  `json:"bid,omitempty"`
}
