package messenger_entities

type Template struct {
	Type    string          `json:"type,omitempty"`
	Payload TemplatePayload `json:"payload,omitempty"`
}

type TemplatePayload struct {
	TemplateType string             `json:"template_type,omitempty"`
	Elements     []TemplateElements `json:"elements,omitempty"`
}

type TemplateElements struct {
	Title    string `json:"title,omitempty"`
	ImageUrl string `json:"image_url,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	// DefaultAction TemplateDefaultAction   `json:"default_action,omitempty"`
	Buttons []TemplateDefaultAction `json:"buttons,omitempty"`
}

type TemplateDefaultAction struct {
	Type               string `json:"type,omitempty"`
	Url                string `json:"url,omitempty"`
	WebviewHeightRatio string `json:"webview_height_ratio,omitempty"`
	Title              string `json:"title,omitempty"`
}

type TemplateButton struct {
	Type    string      `json:"type,omitempty"`
	Url     string      `json:"url,omitempty"`
	Title   string      `json:"title,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}
