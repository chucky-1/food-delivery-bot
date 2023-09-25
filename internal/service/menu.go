package service

import (
	"context"
	"fmt"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
)

type Menu interface {
	GetAllCategories(ctx context.Context) ([]string, error)
	GetActiveDishesByCategory(ctx context.Context, category string) ([]*model.Dish, error)
	GetStoppedDishesByCategory(ctx context.Context, category string) ([]*model.Dish, error)
	GetDish(ctx context.Context, dish string) (*model.Dish, error)
	StopDish(ctx context.Context, dish string) error
	ActivateDish(ctx context.Context, dish string) error
}

type menu struct {
	repo repository.Menu
}

func NewMenu(repo repository.Menu) *menu {
	return &menu{
		repo: repo,
	}
}

func (m *menu) GetAllCategories(ctx context.Context) ([]string, error) {
	categories, err := m.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("getAllCategories: %w", err)
	}
	return categories, nil
}

func (m *menu) GetActiveDishesByCategory(ctx context.Context, category string) ([]*model.Dish, error) {
	dishes, err := m.repo.GetActiveDishesByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("getAllActiveDishesByCategory: %w", err)
	}
	return dishes, nil
}

func (m *menu) GetStoppedDishesByCategory(ctx context.Context, category string) ([]*model.Dish, error) {
	dishes, err := m.repo.GetStoppedDishesByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("getAllStoppedDishesByCategory: %w", err)
	}
	return dishes, nil
}

func (m *menu) GetDish(ctx context.Context, dish string) (*model.Dish, error) {
	d, err := m.repo.GetDish(ctx, dish)
	if err != nil {
		return nil, fmt.Errorf("getDish: %w", err)
	}
	return d, nil
}

func (m *menu) StopDish(ctx context.Context, dish string) error {
	err := m.repo.StopDish(ctx, dish)
	if err != nil {
		return fmt.Errorf("stopDish: %w", err)
	}
	return nil
}

func (m *menu) ActivateDish(ctx context.Context, dish string) error {
	err := m.repo.ActivateDish(ctx, dish)
	if err != nil {
		return fmt.Errorf("activateDish: %w", err)
	}
	return nil
}
