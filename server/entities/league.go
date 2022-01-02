package entities

import "github.com/google/uuid"

type League struct {
	Id      uuid.UUID   `json:"id,omitempty"`
	Name    string      `json:"name,omitempty"`
	Members []uuid.UUID `json:"members,omitempty"`
}
