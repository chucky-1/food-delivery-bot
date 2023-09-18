package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
	"github.com/google/uuid"
)

type Order interface {
	AddDish(ctx context.Context, dish *model.Dish, userTelegramID int64) error
	GetAllDishesByCategory(ctx context.Context, userTelegramID int64) (map[string][]*model.Dish, error)
	GetUserOrdersByOrganizationLunchTime(ctx context.Context, lunchTime string) (map[uuid.UUID]*model.OrderingData, error)
	GetOrganizationsOrdersAmountByDate(ctx context.Context, date time.Time) ([]*model.Statistics, error)
	IsUserHaveAnyOrders(ctx context.Context, userTelegramID int64) (bool, error)
	IsUserHaveConfirmedOrder(ctx context.Context, userTelegramID int64) (bool, error)
	ConfirmOrderByUser(ctx context.Context, userTelegramID int64) error
	ClearOrdersByUser(ctx context.Context, userTelegramID int64, date time.Time) error
	ClearOrdersByUserWithCheckLunchTime(ctx context.Context, userTelegramID int64, date time.Time) error
}

type order struct {
	repo repository.Order
}

func NewOrder(repo repository.Order) *order {
	return &order{
		repo: repo,
	}
}

func (o *order) AddDish(ctx context.Context, dish *model.Dish, userTelegramID int64) error {
	err := o.repo.AddDish(ctx, dish, userTelegramID)
	if err != nil {
		return fmt.Errorf("addDish: %w", err)
	}
	return nil
}

func (o *order) GetAllDishesByCategory(ctx context.Context, userTelegramID int64) (map[string][]*model.Dish, error) {
	dishes, err := o.repo.GetAllDishesByCategory(ctx, userTelegramID)
	if err != nil {
		return nil, fmt.Errorf("getAllDishesByCategory: %w", err)
	}
	return dishes, nil
}

func (o *order) GetUserOrdersByOrganizationLunchTime(ctx context.Context, lunchTime string) (map[uuid.UUID]*model.OrderingData, error) {
	orders, err := o.repo.GetUserOrdersByOrganizationLunchTime(ctx, lunchTime)
	if err != nil {
		return nil, fmt.Errorf("getUserOrdersByOrganizationLunchTime: %w", err)
	}
	return orders, nil
}

func (o *order) GetOrganizationsOrdersAmountByDate(ctx context.Context, date time.Time) ([]*model.Statistics, error) {
	stats, err := o.repo.GetOrganizationsOrdersAmountByDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("getOrganizationsOrdersAmountByDate: %w", err)
	}
	return stats, nil
}

func (o *order) IsUserHaveAnyOrders(ctx context.Context, userTelegramID int64) (bool, error) {
	exist, err := o.repo.IsUserHaveAnyOrders(ctx, userTelegramID)
	if err != nil {
		return false, fmt.Errorf("isUserHaveAnyOrders: %w", err)
	}
	return exist, nil
}

func (o *order) IsUserHaveConfirmedOrder(ctx context.Context, userTelegramID int64) (bool, error) {
	confirmedOrder, err := o.repo.IsUserHaveConfirmedOrder(ctx, userTelegramID)
	if err != nil {
		return false, fmt.Errorf("isUserHaveConfirmedOrder: %w", err)
	}
	return confirmedOrder, nil
}

func (o *order) ConfirmOrderByUser(ctx context.Context, userTelegramID int64) error {
	err := o.repo.ConfirmOrderByUser(ctx, userTelegramID)
	if err != nil {
		return fmt.Errorf("confirmOrderByUser: %w", err)
	}
	return nil
}

func (o *order) ClearOrdersByUser(ctx context.Context, userTelegramID int64, date time.Time) error {
	err := o.repo.ClearOrdersByUser(ctx, userTelegramID, date)
	if err != nil {
		return fmt.Errorf("clearOrderByUser: %w", err)
	}
	return nil
}

func (o *order) ClearOrdersByUserWithCheckLunchTime(ctx context.Context, userTelegramID int64, date time.Time) error {
	err := o.repo.ClearOrdersByUserWithCheckLunchTime(ctx, userTelegramID, date)
	if err != nil {
		return fmt.Errorf("clearOrdersByUserWithCheckLunchTime: %w", err)
	}
	return nil
}
