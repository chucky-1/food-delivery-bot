package service

import (
	"context"
	"fmt"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
	"github.com/google/uuid"
)

type Auth interface {
	Register(ctx context.Context, u *model.TelegramUser) error
}

type auth struct {
	userRepo     repository.User
	telegramRepo repository.Telegram
	orgRepo      repository.Organization
	transactor   repository.Transactor
}

func NewAuth(userRepo repository.User, telegramRepo repository.Telegram, orgRepo repository.Organization, transactor repository.Transactor) *auth {
	return &auth{
		userRepo:     userRepo,
		telegramRepo: telegramRepo,
		orgRepo:      orgRepo,
		transactor:   transactor,
	}
}

func (a *auth) Register(ctx context.Context, telegramUser *model.TelegramUser) error {
	err := a.transactor.Transact(ctx, func(ctx context.Context) error {
		err := a.telegramRepo.AddUser(ctx, telegramUser)
		if err != nil {
			return fmt.Errorf("addUser: %w", err)
		}
		err = a.userRepo.Add(ctx, &model.User{
			ID:             uuid.New(),
			TelegramID:     telegramUser.ID,
			OrganizationID: uuid.Nil,
		})
		if err != nil {
			return fmt.Errorf("add: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	return nil
}
