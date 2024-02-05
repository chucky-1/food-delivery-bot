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
	UpdateFirstName(ctx context.Context, telegramUserID int, firstName string) error
	UpdateLastName(ctx context.Context, telegramUserID int, lastName string) error
	UpdateMiddleName(ctx context.Context, telegramUserID int, middleName string) error
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

func (a *auth) UpdateFirstName(ctx context.Context, telegramUserID int, firstName string) error {
	err := a.userRepo.UpdateFirstName(ctx, telegramUserID, firstName)
	if err != nil {
		return fmt.Errorf("updateFirstName: %w", err)
	}
	return nil
}

func (a *auth) UpdateLastName(ctx context.Context, telegramUserID int, lastName string) error {
	err := a.userRepo.UpdateLastName(ctx, telegramUserID, lastName)
	if err != nil {
		return fmt.Errorf("updateLastName: %w", err)
	}
	return nil
}

func (a *auth) UpdateMiddleName(ctx context.Context, telegramUserID int, middleName string) error {
	err := a.userRepo.UpdateMiddleName(ctx, telegramUserID, middleName)
	if err != nil {
		return fmt.Errorf("updateMiddleName: %w", err)
	}
	return nil
}
