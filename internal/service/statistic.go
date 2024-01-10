package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
	"github.com/google/uuid"
)

type Statistics interface {
	Get(ctx context.Context, from, to time.Time) (map[uuid.UUID]*model.Statistic, error)
}

type statistics struct {
	order      repository.Order
	transactor repository.Transactor
}

func NewStatistics(order repository.Order, transactor repository.Transactor) *statistics {
	return &statistics{
		order:      order,
		transactor: transactor,
	}
}

func (s *statistics) Get(ctx context.Context, from, to time.Time) (map[uuid.UUID]*model.Statistic, error) {
	stats, err := s.order.GetOrdersAmount(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("getOrdersAmount: %w", err)
	}
	return stats, nil
}
