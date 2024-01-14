package service

import (
	"context"
	"fmt"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
	"github.com/google/uuid"
)

type Organization interface {
	Add(ctx context.Context, org *model.Organization) error
	Join(ctx context.Context, organizationID uuid.UUID, userTelegramID int64) error
	UpdateAddress(ctx context.Context, id uuid.UUID, address string) error
}

type organization struct {
	repo repository.Organization
}

func NewOrganization(repo repository.Organization) *organization {
	return &organization{
		repo: repo,
	}
}

func (o *organization) Add(ctx context.Context, org *model.Organization) error {
	if err := o.repo.Add(ctx, org); err != nil {
		return fmt.Errorf("add: %w", err)
	}
	return nil
}

func (o *organization) Join(ctx context.Context, organizationID uuid.UUID, userTelegramID int64) error {
	if err := o.repo.Join(ctx, organizationID, userTelegramID); err != nil {
		return fmt.Errorf("join: %w", err)
	}
	return nil
}

func (o *organization) UpdateAddress(ctx context.Context, id uuid.UUID, address string) error {
	err := o.repo.UpdateAddress(ctx, id, address)
	if err != nil {
		return fmt.Errorf("updateAddress: %w", err)
	}
	return nil
}
