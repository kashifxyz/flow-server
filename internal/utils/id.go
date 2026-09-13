package utils

import (
	"net/http"

	"github.com/google/uuid"
)

func NewID() (uuid.UUID, error) {
	return uuid.NewV7()
}

func PathID(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func MustID() uuid.UUID {
	id, err := NewID()
	if err != nil {
		panic(err)
	}
	return id
}

func MustID() uuid.UUID {
	id, err := NewID()
	if err != nil {
		panic(err)
	}
	return id
}
