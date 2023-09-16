package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
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
}

type order struct {
	tr       *transactor
	timezone time.Duration
}

func NewOrder(tr *transactor, timezone time.Duration) *order {
	return &order{
		tr:       tr,
		timezone: timezone,
	}
}

func (o *order) AddDish(ctx context.Context, dish *model.Dish, userTelegramID int64) error {
	query := `INSERT INTO internal.orders (date, user_telegram_id, dish_name, dish_price, category) VALUES ($1,$2,$3,$4,$5)`
	date := time.Now().UTC().Add(o.timezone)
	_, err := o.tr.extractTx(ctx).Exec(ctx, query, date, userTelegramID, dish.Name, dish.Price, dish.Category)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (o *order) GetAllDishesByCategory(ctx context.Context, userTelegramID int64) (map[string][]*model.Dish, error) {
	query := `SELECT dish_name, dish_price, category FROM internal.orders WHERE user_telegram_id = $1`
	rows, err := o.tr.extractTx(ctx).Query(ctx, query, userTelegramID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	dishes := make(map[string][]*model.Dish)
	for rows.Next() {
		var dish model.Dish
		err = rows.Scan(&dish.Name, &dish.Price, &dish.Category)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		dishes[dish.Category] = append(dishes[dish.Category], &dish)
	}
	return dishes, nil
}

func (o *order) GetUserOrdersByOrganizationLunchTime(ctx context.Context, lunchTime string) (map[uuid.UUID]*model.OrderingData, error) {
	query := `
		SELECT org.id, org.name, org.address, o.dish_name, o.dish_price, o.category, count(1)
		FROM internal.orders o
		LEFT JOIN internal.users u ON u.telegram_id = o.user_telegram_id
		LEFT JOIN internal.organizations org ON org.id = u.organization_id
		WHERE o.confirmed = true AND org.lunch_time = $1 AND date = $2
		GROUP BY org.id, o.dish_name, o.dish_price, o.category`
	date := time.Now().UTC().Add(o.timezone)

	rows, err := o.tr.extractTx(ctx).Query(ctx, query, lunchTime, date)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	res := make(map[uuid.UUID]*model.OrderingData)
	for rows.Next() {
		var (
			orgID      uuid.UUID
			orgName    string
			orgAddress string
			dishName   string
			dishPrice  float32
			category   string
			count      int
		)
		err = rows.Scan(&orgID, &orgName, &orgAddress, &dishName, &dishPrice, &category, &count)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		data, ok := res[orgID]
		if !ok {
			data = &model.OrderingData{
				OrganizationName:    orgName,
				OrganizationAddress: orgAddress,
				DishesByCategories:  make(map[string][]*model.DishWithCount),
			}
			res[orgID] = data
		}
		data.DishesByCategories[category] = append(data.DishesByCategories[category], &model.DishWithCount{
			Dish: &model.Dish{
				Name:  dishName,
				Price: dishPrice,
			},
			Count: count,
		})
	}
	return res, nil
}

func (o *order) GetOrganizationsOrdersAmountByDate(ctx context.Context, date time.Time) ([]*model.Statistics, error) {
	query := `
		SELECT org.id, sum(o.dish_price)
		FROM internal.orders o
		LEFT JOIN internal.users u ON u.telegram_id = o.user_telegram_id
		LEFT JOIN internal.organizations org ON org.id = u.organization_id
		WHERE o.confirmed = true AND o.date = $1
		GROUP BY org.id`

	rows, err := o.tr.extractTx(ctx).Query(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	res := make([]*model.Statistics, 0)
	for rows.Next() {
		var st model.Statistics
		err = rows.Scan(&st.OrganizationID, &st.OrdersAmount)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		res = append(res, &st)
	}
	return res, nil
}

func (o *order) IsUserHaveAnyOrders(ctx context.Context, userTelegramID int64) (bool, error) {
	query := `SELECT EXISTS (
    SELECT 1
    FROM internal.orders
    WHERE user_telegram_id = $1)`
	var exist bool
	err := o.tr.extractTx(ctx).QueryRow(ctx, query, userTelegramID).Scan(&exist)
	if err != nil {
		return false, fmt.Errorf("queryRow: %w", err)
	}
	return exist, nil
}

func (o *order) IsUserHaveConfirmedOrder(ctx context.Context, userTelegramID int64) (bool, error) {
	query := `SELECT EXISTS (
    SELECT 1
    FROM internal.orders
    WHERE user_telegram_id = $1
    AND confirmed = true)`
	var exist bool
	err := o.tr.extractTx(ctx).QueryRow(ctx, query, userTelegramID).Scan(&exist)
	if err != nil {
		return false, fmt.Errorf("queryRow: %w", err)
	}
	return exist, nil
}

func (o *order) ConfirmOrderByUser(ctx context.Context, userTelegramID int64) error {
	query := `UPDATE internal.orders SET confirmed = true WHERE user_telegram_id = $1`
	_, err := o.tr.extractTx(ctx).Exec(ctx, query, userTelegramID)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (o *order) ClearOrdersByUser(ctx context.Context, userTelegramID int64, date time.Time) error {
	query := `DELETE FROM internal.orders WHERE user_telegram_id = $1 AND date = $2`
	_, err := o.tr.extractTx(ctx).Exec(ctx, query, userTelegramID, date)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}
