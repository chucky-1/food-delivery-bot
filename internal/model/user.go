package model

import "github.com/google/uuid"

type User struct {
	ID             uuid.UUID
	TelegramID     int64
	OrganizationID uuid.UUID
}
