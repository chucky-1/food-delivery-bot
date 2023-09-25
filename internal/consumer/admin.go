package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

var (
	welcomeAdminMessage = "/all_stopped_dishes - показать все блюда на стопе\n\n" +
		"/all_active_dishes - показать доступные для заказа блюда\n\n" +
		"Что бы поставить блюдо на стоп, жмём /all_active_dishes, выбираем блюдо\n\n" +
		"Что бы снять блюдо со стопа, жмём /all_stopped_dishes, выбираем блюдо\n\n" +
		"/info - показать это сообщение (можно ввести эту команду руками, когда это сообщение потеряется в куче других сообщений)"
)

const (
	info              = "info"
	allActivateDishes = "all_active_dishes"
	allStoppedDishes  = "all_stopped_dishes"
)

type Admin struct {
	bot         *tgbotapi.BotAPI
	updatesChan chan tgbotapi.Update
	menu        service.Menu
	adminID     int64

	// if true - state to activate dishes
	state bool
}

func NewAdmin(bot *tgbotapi.BotAPI, updatesChan chan tgbotapi.Update, menu service.Menu, adminID int64) *Admin {
	return &Admin{
		bot:         bot,
		updatesChan: updatesChan,
		menu:        menu,
		adminID:     adminID,
	}
}

func (a *Admin) Consume(ctx context.Context) {
	logrus.Info("admin consumer started")
	err := a.sendWelcomeMessage()
	if err != nil {
		logrus.Errorf("admin: %s", err.Error())
		return
	}
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("admin consumer stopped: %s", ctx.Err().Error())
			return
		case update := <-a.updatesChan:
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case info:
					err = a.sendWelcomeMessage()
					if err != nil {
						logrus.Errorf("admin: info: %s", err.Error())
						continue
					}
				case allActivateDishes:
					a.state = false

					newCtx, cancel := context.WithTimeout(ctx, time.Minute)
					err = a.sendCategories(newCtx, update.Message.Chat.ID)
					if err != nil {
						cancel()
						logrus.Errorf("admin: allActivateDishes: %s", err.Error())
						continue
					}
					cancel()
					continue

				case allStoppedDishes:
					a.state = true

					newCtx, cancel := context.WithTimeout(ctx, time.Minute)
					err = a.sendCategories(newCtx, update.Message.Chat.ID)
					if err != nil {
						cancel()
						logrus.Errorf("admin: allActivateDishes: %s", err.Error())
						continue
					}
					cancel()
					continue
				}
			} else {
				switch update.Message.Text {
				case model.Soups:
					err = a.sendDishes(ctx, update.Message.Chat.ID, model.Soups, a.state)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case model.Salads:
					err = a.sendDishes(ctx, update.Message.Chat.ID, model.Salads, a.state)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case model.MainCourse:
					err = a.sendDishes(ctx, update.Message.Chat.ID, model.MainCourse, a.state)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case model.Desserts:
					err = a.sendDishes(ctx, update.Message.Chat.ID, model.Desserts, a.state)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case model.Drinks:
					err = a.sendDishes(ctx, update.Message.Chat.ID, model.Drinks, a.state)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case goBackToMenu, "Меню":
					err = a.sendCategories(ctx, update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("sendMenu: %s", err.Error())
						continue
					}
					continue
				}

				newCtx, cancel := context.WithTimeout(ctx, time.Minute)
				dish, err := a.menu.GetDish(newCtx, update.Message.Text)
				if err != nil {
					logrus.Errorf("admin: %s", err.Error())
					cancel()
					continue
				}
				if dish != nil {
					switch a.state {
					case true:
						err = a.menu.ActivateDish(ctx, dish.String())
						if err != nil {
							logrus.Errorf("admin: %s", err.Error())
							cancel()
							continue
						}
					case false:
						err = a.menu.StopDish(ctx, dish.String())
						if err != nil {
							logrus.Errorf("admin: %s", err.Error())
							cancel()
							continue
						}
					}

					err = a.sendDishes(newCtx, update.Message.Chat.ID, dish.Category, a.state)
					if err != nil {
						logrus.Error(err.Error())
						cancel()
						continue
					}
					cancel()
					continue
				}
				cancel()
				continue
			}
		}
	}
}

func (a *Admin) sendWelcomeMessage() error {
	msg := tgbotapi.NewMessage(a.adminID, welcomeAdminMessage)
	_, err := a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (a *Admin) sendCategories(ctx context.Context, chatID int64) error {
	categories, err := a.menu.GetAllCategories(ctx)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, "Меню")
	var buttons [][]tgbotapi.KeyboardButton
	for _, category := range categories {
		but := tgbotapi.NewKeyboardButton(category)
		row := tgbotapi.NewKeyboardButtonRow(but)
		buttons = append(buttons, row)
	}
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons...)
	_, err = a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (a *Admin) sendDishes(ctx context.Context, chatID int64, category string, active bool) error {
	var (
		dishes []*model.Dish
		err    error
	)
	switch active {
	case false:
		dishes, err = a.menu.GetActiveDishesByCategory(ctx, category)
		if err != nil {
			return err
		}
	case true:
		dishes, err = a.menu.GetStoppedDishesByCategory(ctx, category)
		if err != nil {
			return err
		}
	}

	msg := tgbotapi.NewMessage(chatID, category)
	var buttons [][]tgbotapi.KeyboardButton
	for _, dish := range dishes {
		but := tgbotapi.NewKeyboardButton(dish.String())
		row := tgbotapi.NewKeyboardButtonRow(but)
		buttons = append(buttons, row)
	}
	but := tgbotapi.NewKeyboardButton(goBackToMenu)
	row := tgbotapi.NewKeyboardButtonRow(but)
	buttons = append(buttons, row)

	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons...)
	_, err = a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}
