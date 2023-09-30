package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/google/uuid"
)

type User interface {
	Add(ctx context.Context, usr *model.User) error
}

type user struct {
	tr *transactor
}

func NewUser(tr *transactor) *user {
	return &user{
		tr: tr,
	}
}

func (u *user) Add(ctx context.Context, usr *model.User) error {
	var orgID sql.NullString
	if usr.OrganizationID != uuid.Nil {
		orgID.Valid = true
		orgID.String = usr.OrganizationID.String()
	}
	query := `INSERT INTO internal.users (id, telegram_id, organization_id) VALUES ($1,$2,$3)`
	_, err := u.tr.extractTx(ctx).Exec(ctx, query, usr.ID, usr.TelegramID, orgID)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}
