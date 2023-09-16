package model

import "fmt"

const (
	Soups      = "Супы"
	Salads     = "Салаты"
	MainCourse = "Основные блюда"
	Desserts   = "Десерты"
	Drinks     = "Напитки"
)

type Dish struct {
	Name     string
	Price    float32
	Category string
}

func (d *Dish) String() string {
	return fmt.Sprintf("%s - %.2f", d.Name, d.Price)
}
