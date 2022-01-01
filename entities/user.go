package entities

import "github.com/google/uuid"

type User struct {
	Id   uuid.UUID `json:"id,omitempty"`
	Name string    `json:"name,omitempty"`
}
