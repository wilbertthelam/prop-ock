package messenger_entities

type WebhookBody struct {
	Object string
	Entry  []WebhookEntry
}

type WebhookEntry struct {
	Messaging []WebhookEvent
}

type WebhookEvent struct {
	Sender   WebhookEventSender
	Message  WebhookMessageEvent  `json:"message,omitempty"`
	Postback WebhookPostbackEvent `json:"postback,omitempty"`
	Read     WebhookReadEvent     `json:"read,omitempty"`
}

type WebhookMessageEvent struct {
	Sender    Id
	Recipient Id
	Timestamp int64
	Message   WebhookMessage
}

type WebhookMessage struct {
	Mid  string
	Text string
}

type WebhookPostbackEvent struct {
	Sender    Id
	Recipient Id
	Timestamp int64
	Postback  WebhookPostback
}

type WebhookPostback struct {
	Title   string
	Payload string
}

type WebhookReadEvent struct{}

type WebhookEventSender struct {
	Id string
}
