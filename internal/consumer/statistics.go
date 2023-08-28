package consumer

import (
	"context"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/service"
	"github.com/sirupsen/logrus"
)

type Statistics struct {
	statistics service.Statistics
	timezone   time.Duration
}

func NewStatistics(statistics service.Statistics, timezone time.Duration) *Statistics {
	return &Statistics{
		statistics: statistics,
		timezone:   timezone,
	}
}

func (s *Statistics) Consume(ctx context.Context) {
	logrus.Info("statistics consumer started")
	s.waitToCreateTicker(ctx)
	logrus.Info("statistics consumer is ready to create ticker: %s", time.Now().UTC().String())
	t := time.NewTicker(time.Hour)
	for {
		select {
		case <-ctx.Done():
			logrus.Info("statistics consumer stopped: %s", ctx.Err().Error())
			t.Stop()
			return
		case <-t.C:
			if time.Now().UTC().Add(s.timezone).Hour() != 0 {
				continue
			}
			logrus.Infof("statistics consumer started executing: %s", time.Now().UTC().String())

			newCtx, cancel := context.WithTimeout(ctx, time.Minute)
			err := s.statistics.AddByDate(newCtx, time.Now().UTC().Add(s.timezone).Add(-time.Hour))
			if err != nil {
				logrus.Errorf("statistics consumer: %s", err.Error())
				cancel()
				continue
			}
			cancel()
		}
	}
}

func (s *Statistics) waitToCreateTicker(ctx context.Context) {
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
			if time.Now().Minute() == 0 {
				t.Stop()
				return
			}
		}
	}
}
