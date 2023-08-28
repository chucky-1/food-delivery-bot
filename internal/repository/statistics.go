package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

type Statistics interface {
	Add(ctx context.Context, stats []*model.Statistics, date time.Time) error
	GetByDate(ctx context.Context, date time.Time) ([]*model.Statistics, error)
	GetByDatesInterval(ctx context.Context, from, to time.Time) ([]*model.Statistics, error)
}

type statistics struct {
	tr *transactor
}

func NewStatistics(tr *transactor) *statistics {
	return &statistics{
		tr: tr,
	}
}

func (s *statistics) Add(ctx context.Context, stats []*model.Statistics, date time.Time) error {
	query := `INSERT INTO internal.statistics (date, organization_id, order_amount) VALUES ($1,$2,$3)`

	batch := pgx.Batch{}
	for _, st := range stats {
		batch.Queue(query, date, st.OrganizationID, st.OrdersAmount)
	}
	res := s.tr.extractTx(ctx).SendBatch(ctx, &batch)

	for i := 0; i < batch.Len(); i++ {
		_, err := res.Exec()
		if err != nil {
			return fmt.Errorf("exec: %w", err)
		}
	}

	err := res.Close()
	if err != nil {
		logrus.Errorf("statistics couldn't close batch: %s", err.Error())
	}

	return nil
}

func (s *statistics) GetByDate(ctx context.Context, date time.Time) ([]*model.Statistics, error) {
	query := `
		SELECT o.name, order_amount 
		FROM internal.statistics
		LEFT JOIN internal.organizations o ON o.id = statistics.organization_id
		WHERE date = $1`
	rows, err := s.tr.extractTx(ctx).Query(ctx, query, date)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	stats := make([]*model.Statistics, 0)
	for rows.Next() {
		var st model.Statistics
		err = rows.Scan(&st.OrganizationName, &st.OrdersAmount)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		stats = append(stats, &st)
	}
	return stats, nil
}

func (s *statistics) GetByDatesInterval(ctx context.Context, from, to time.Time) ([]*model.Statistics, error) {
	query := `
		SELECT o.name, sum(s.order_amount) 
		FROM internal.organizations o
		LEFT JOIN internal.statistics s ON s.organization_id = o.id
		WHERE date >= $1 and date <= $2
		GROUP BY o.id`

	rows, err := s.tr.extractTx(ctx).Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	stats := make([]*model.Statistics, 0)
	for rows.Next() {
		var st model.Statistics
		err = rows.Scan(&st.OrganizationName, &st.OrdersAmount)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		stats = append(stats, &st)
	}
	return stats, nil
}
