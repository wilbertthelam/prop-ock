package messenger_entities

type WebhookBody struct {
	Object string
	Entry  []WebhookEntry
}

type WebhookEntry struct {
	Messaging []WebhookEvent
}

type WebhookEvent struct {
	Sender    Id
	Timestamp int64
	Message   WebhookMessage  `json:"message,omitempty"`
	Postback  WebhookPostback `json:"postback,omitempty"`
	Read      WebhookRead     `json:"read,omitempty"`
}

type WebhookMessage struct {
	Mid  string
	Text string
}

type WebhookPostback struct {
	Title   string
	Payload string
}

type WebhookRead struct{}
