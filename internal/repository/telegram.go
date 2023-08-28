package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
)

type Telegram interface {
	AddUser(ctx context.Context, u *model.TelegramUser) error
	GetUsersByLunchTimes(ctx context.Context, lunchTimes []string) (map[time.Duration][]*model.TelegramUser, error)
}

type telegram struct {
	tr *transactor
}

func NewTelegram(tr *transactor) *telegram {
	return &telegram{
		tr: tr,
	}
}

func (t *telegram) AddUser(ctx context.Context, u *model.TelegramUser) error {
	query := `INSERT INTO telegram.users (id, chat_id, first_name) VALUES ($1,$2,$3)`
	_, err := t.tr.extractTx(ctx).Exec(ctx, query, u.ID, u.ChatID, u.FirstName)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	return nil
}

func (t *telegram) GetUsersByLunchTimes(ctx context.Context, lunchTimes []string) (map[time.Duration][]*model.TelegramUser, error) {
	query := `SELECT t.id AS telegram_user_id, t.chat_id, t.first_name, io.lunch_time
	FROM telegram.users AS t
	JOIN internal.users AS iu ON t.id = iu.telegram_id
	JOIN internal.organizations AS io ON iu.organization_id = io.id
	WHERE io.lunch_time = ANY($1::interval[])
	AND NOT EXISTS (
    	SELECT 1
    	FROM internal.orders AS o
    	WHERE o.user_telegram_id = t.id
    	AND o.confirmed = true
    	)`
	rows, err := t.tr.extractTx(ctx).Query(ctx, query, lunchTimes)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	telegramUsers := make(map[time.Duration][]*model.TelegramUser)
	for rows.Next() {
		var (
			telegramUser model.TelegramUser
			lunchTime    time.Duration
		)
		err = rows.Scan(&telegramUser.ID, &telegramUser.ChatID, &telegramUser.FirstName, &lunchTime)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		telegramUsers[lunchTime] = append(telegramUsers[lunchTime], &telegramUser)
	}
	return telegramUsers, nil
}
