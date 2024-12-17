package api

import (
	"github.com/google/uuid"
)

type Client struct {
	ClientToken string
}

type ObjectID struct {
	ID uuid.UUID `json:"id"`
}
