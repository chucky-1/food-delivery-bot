package model

type DishWithCount struct {
	*Dish
	Count int
}

type OrderingData struct {
	OrganizationName    string
	OrganizationAddress string
	DishesByCategories  map[string][]*DishWithCount
}
