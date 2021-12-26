package messenger_entities

type Id struct {
	Id string `json:"id,omitempty"`
}

type SendEvent struct {
	Recipient Id          `json:"recipient,omitempty"`
	Message   SendMessage `json:"message,omitempty"`
}

type SendMessage struct {
	Attachment Template `json:"attachment,omitempty"`
}
