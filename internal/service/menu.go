package service

import (
	"context"
	"fmt"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
)

type Menu interface {
	GetAllCategories(ctx context.Context) ([]string, error)
	GetAllDishesByCategory(ctx context.Context, category string) ([]*model.Dish, error)
	GetDish(ctx context.Context, dish string) (*model.Dish, error)
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

func (m *menu) GetAllDishesByCategory(ctx context.Context, category string) ([]*model.Dish, error) {
	dishes, err := m.repo.GetAllDishesByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("getAllDishesByCategory: %w", err)
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
