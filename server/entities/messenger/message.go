package messenger_entities

type Id struct {
	Id string `json:"id,omitempty"`
}

type SendEvent struct {
	Recipient Id          `json:"recipient,omitempty"`
	Message   SendMessage `json:"message,omitempty"`
	Tag       string      `json:"tag,omitempty"`
}

type SendMessage struct {
	Attachment Template `json:"attachment,omitempty"`
}

type SendEventResponse struct {
	Error SendEventResponseError `json:"error,omitempty"`
}

type SendEventResponseError struct {
	Message      string `json:"message,omitempty"`
	Type         string `json:"type,omitempty"`
	Code         int64  `json:"code,omitempty"`
	ErrorSubcode int64  `json:"error_subcode,omitempty"`
	FBTraceId    string `json:"fbtrace_id,omitempty"`
}
