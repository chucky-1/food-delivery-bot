package repository

import (
	"context"

	"github.com/chucky-1/food-delivery-bot/internal/model"
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
	categories                []string
	allDishesByCategories     map[string][]*model.Dish
	activeDishesByCategories  map[string][]*model.Dish
	stoppedDishesByCategories map[string][]*model.Dish
	allDishes                 map[string]*model.Dish
}

func NewMenu(categories []string, allDishesByCategories, activeDishesByCategories, stoppedDishesByCategories map[string][]*model.Dish,
	allDishes map[string]*model.Dish) *menu {
	return &menu{
		categories:                categories,
		allDishesByCategories:     allDishesByCategories,
		activeDishesByCategories:  activeDishesByCategories,
		stoppedDishesByCategories: stoppedDishesByCategories,
		allDishes:                 allDishes,
	}
}

func (m *menu) GetAllCategories(_ context.Context) ([]string, error) {
	return m.categories, nil
}

func (m *menu) GetActiveDishesByCategory(_ context.Context, category string) ([]*model.Dish, error) {
	return m.activeDishesByCategories[category], nil
}

func (m *menu) GetStoppedDishesByCategory(ctx context.Context, category string) ([]*model.Dish, error) {
	return m.stoppedDishesByCategories[category], nil
}

func (m *menu) GetDish(_ context.Context, dish string) (*model.Dish, error) {
	return m.getDish(dish)
}

func (m *menu) StopDish(_ context.Context, dish string) error {
	myDish, err := m.getDish(dish)
	if err != nil {
		return err
	}
	m.changeDishState(myDish, true)

	activeDishes := m.updateActiveDishes()
	m.activeDishesByCategories = activeDishes

	stoppedDishes := m.updateStoppedDishes()
	m.stoppedDishesByCategories = stoppedDishes

	return nil
}

func (m *menu) ActivateDish(_ context.Context, dish string) error {
	myDish, err := m.getDish(dish)
	if err != nil {
		return err
	}
	m.changeDishState(myDish, false)

	activeDishes := m.updateActiveDishes()
	m.activeDishesByCategories = activeDishes

	stoppedDishes := m.updateStoppedDishes()
	m.stoppedDishesByCategories = stoppedDishes
	return nil
}

func (m *menu) getDish(dish string) (*model.Dish, error) {
	d, ok := m.allDishes[dish]
	if !ok {
		return nil, nil
	}
	return d, nil
}

func (m *menu) updateActiveDishes() map[string][]*model.Dish {
	dishes := make(map[string][]*model.Dish)
	for category, ds := range m.allDishesByCategories {
		for _, dish := range ds {
			if dish.Stop {
				continue
			}
			dishes[category] = append(dishes[category], dish)
		}
	}
	return dishes
}

func (m *menu) updateStoppedDishes() map[string][]*model.Dish {
	dishes := make(map[string][]*model.Dish)
	for category, ds := range m.allDishesByCategories {
		for _, dish := range ds {
			if !dish.Stop {
				continue
			}
			dishes[category] = append(dishes[category], dish)
		}
	}
	return dishes
}

func (m *menu) changeDishState(dish *model.Dish, state bool) {
	dishes := m.allDishesByCategories[dish.Category]
	for idx, resDish := range dishes {
		if resDish.Name == dish.Name {
			dishes[idx].Stop = state
			break
		}
	}
}
