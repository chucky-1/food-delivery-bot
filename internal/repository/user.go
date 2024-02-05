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
	UpdateFirstName(ctx context.Context, telegramUserID int, firstName string) error
	UpdateLastName(ctx context.Context, telegramUserID int, lastName string) error
	UpdateMiddleName(ctx context.Context, telegramUserID int, middleName string) error
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
	query := `INSERT INTO internal.users (id, telegram_id, organization_id, first_name, last_name, middle_name) VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := u.tr.extractTx(ctx).Exec(ctx, query, usr.ID, usr.TelegramID, orgID, usr.FirstName, usr.LastName, usr.MiddleName)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (u *user) UpdateFirstName(ctx context.Context, telegramUserID int, firstName string) error {
	query := `UPDATE internal.users SET first_name=$1 WHERE telegram_id=$2`
	_, err := u.tr.extractTx(ctx).Exec(ctx, query, firstName, telegramUserID)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (u *user) UpdateLastName(ctx context.Context, telegramUserID int, lastName string) error {
	query := `UPDATE internal.users SET last_name=$1 WHERE telegram_id=$2`
	_, err := u.tr.extractTx(ctx).Exec(ctx, query, lastName, telegramUserID)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (u *user) UpdateMiddleName(ctx context.Context, telegramUserID int, middleName string) error {
	query := `UPDATE internal.users SET middle_name=$1 WHERE telegram_id=$2`
	_, err := u.tr.extractTx(ctx).Exec(ctx, query, middleName, telegramUserID)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}
