package model

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID
	Name      string
	OwnerID   uuid.UUID
	LunchTime time.Duration
}
