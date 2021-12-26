package entities

import "github.com/google/uuid"

type League struct {
	Id      uuid.UUID
	Name    string
	Members []uuid.UUID
}
