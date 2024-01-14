package repository

import (
	"context"
	"fmt"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/google/uuid"
)

type Organization interface {
	Add(ctx context.Context, org *model.Organization) error
	Join(ctx context.Context, organizationID uuid.UUID, userTelegramID int64) error
	UpdateAddress(ctx context.Context, id uuid.UUID, address string) error
}

type organization struct {
	tr *transactor
}

func NewOrganization(tr *transactor) *organization {
	return &organization{
		tr: tr,
	}
}

func (o *organization) Add(ctx context.Context, org *model.Organization) error {
	query := `INSERT INTO internal.organizations (id, name, lunch_time) VALUES ($1,$2,$3)`
	_, err := o.tr.extractTx(ctx).Exec(ctx, query, org.ID, org.Name, org.LunchTime)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (o *organization) Join(ctx context.Context, organizationID uuid.UUID, userTelegramID int64) error {
	query := `UPDATE internal.users SET organization_id = $1 WHERE telegram_id = $2`
	_, err := o.tr.extractTx(ctx).Exec(ctx, query, organizationID, userTelegramID)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (o *organization) UpdateAddress(ctx context.Context, id uuid.UUID, address string) error {
	query := `UPDATE internal.organizations
	SET address = $1
    WHERE id = $2`

	_, err := o.tr.extractTx(ctx).Exec(ctx, query, address, id)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}
