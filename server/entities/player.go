package entities

type Player struct {
	Id       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Image    string `json:"image,omitempty"`
	Team     string `json:"team,omitempty"`
	Position string `json:"position,omitempty"`
}
