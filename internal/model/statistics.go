package model

import (
	"github.com/google/uuid"
)

type Statistics struct {
	OrganizationID   uuid.UUID
	OrganizationName string
	OrdersAmount     float32
}
