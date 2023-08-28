package repository

import (
	"context"

	"github.com/chucky-1/food-delivery-bot/internal/model"
)

type Menu interface {
	GetAllCategories(ctx context.Context) ([]string, error)
	GetAllDishesByCategory(ctx context.Context, category string) ([]*model.Dish, error)
	GetDish(ctx context.Context, dish string) (*model.Dish, error)
}

type menu struct {
	categories         []string
	dishesByCategories map[string][]*model.Dish
	allDishes          map[string]*model.Dish
}

func NewMenu(categories []string, dishesByCategories map[string][]*model.Dish, allDishes map[string]*model.Dish) *menu {
	return &menu{
		categories:         categories,
		dishesByCategories: dishesByCategories,
		allDishes:          allDishes,
	}
}

func (m *menu) GetAllCategories(_ context.Context) ([]string, error) {
	return m.categories, nil
}

func (m *menu) GetAllDishesByCategory(_ context.Context, category string) ([]*model.Dish, error) {
	return m.dishesByCategories[category], nil
}

func (m *menu) GetDish(_ context.Context, dish string) (*model.Dish, error) {
	d, ok := m.allDishes[dish]
	if !ok {
		return nil, nil
	}
	return d, nil
}
