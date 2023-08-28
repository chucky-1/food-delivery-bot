package service

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
)

type Telegram interface {
	GetUsersByLunchTimes(ctx context.Context, lunchTimes []string) (map[time.Duration][]*model.TelegramUser, error)
}

type telegram struct {
	repo repository.Telegram
}

func NewTelegram(repo repository.Telegram) *telegram {
	return &telegram{
		repo: repo,
	}
}

func (t *telegram) GetUsersByLunchTimes(ctx context.Context, lunchTimes []string) (map[time.Duration][]*model.TelegramUser, error) {
	telegramUsers, err := t.repo.GetUsersByLunchTimes(ctx, lunchTimes)
	if err != nil {
		return nil, fmt.Errorf("getUsersByLunchTimes: %w", err)
	}
	return telegramUsers, nil
}
