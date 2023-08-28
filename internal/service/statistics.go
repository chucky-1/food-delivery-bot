package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
)

type Statistics interface {
	AddByDate(ctx context.Context, date time.Time) error
	GetByDate(ctx context.Context, date time.Time) ([]*model.Statistics, error)
	GetByDatesInterval(ctx context.Context, from, to time.Time) ([]*model.Statistics, error)
}

type statistics struct {
	repo       repository.Statistics
	order      repository.Order
	transactor repository.Transactor
}

func NewStatistics(repo repository.Statistics, order repository.Order, transactor repository.Transactor) *statistics {
	return &statistics{
		repo:       repo,
		order:      order,
		transactor: transactor,
	}
}

func (s *statistics) AddByDate(ctx context.Context, date time.Time) error {
	err := s.transactor.Transact(ctx, func(ctx context.Context) error {
		stats, err := s.order.GetOrganizationsOrdersAmountByDate(ctx, date)
		if err != nil {
			return fmt.Errorf("getOrganizationsOrdersAmountByDate: %w", err)
		}
		err = s.repo.Add(ctx, stats, date)
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("add: %w", err)
	}
	return nil
}

func (s *statistics) GetByDate(ctx context.Context, date time.Time) ([]*model.Statistics, error) {
	stats, err := s.repo.GetByDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("getByDate: %w", err)
	}
	return stats, nil
}

func (s *statistics) GetByDatesInterval(ctx context.Context, from, to time.Time) ([]*model.Statistics, error) {
	stats, err := s.repo.GetByDatesInterval(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("getByDate: %w", err)
	}
	return stats, nil
}
