package consumer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/service"
	"github.com/chucky-1/food-delivery-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	welcomeAdminMessage = "/all_stopped_dishes - показать все блюда на стопе\n\n" +
		"/all_active_dishes - показать доступные для заказа блюда\n\n" +
		"Что бы поставить блюдо на стоп, жмём /all_active_dishes, выбираем блюдо\n\n" +
		"Что бы снять блюдо со стопа, жмём /all_stopped_dishes, выбираем блюдо\n\n" +
		"Создать организацию /create_organization\n\n" +
		"/info - показать это сообщение (можно ввести эту команду руками, когда это сообщение потеряется в куче других сообщений)"
	createOrganization = "Отправьте сообщение в следующем формате: \n\n" +
		"Название организации 12:30\n\n" +
		"Где 12:30 - это время, к которому нужно осуществить доставку"
	successfulOrganizationRegistered = "Организация успешно создана: %s\n\n" +
		"Чтобы присоединиться к ней, потребуется уникальный идентификатор (ID):"
	addAddressAfterCreateOrganizationMessage = "Добавьте адрес организации, куда доставлять обеды. " +
		"Пример:\n\n" +
		"ул. Толбухина 18/2"
	successfulAddAddress = "Адрес организации успешно добавлен"
)

const (
	info                      = "info"
	allActivateDishes         = "all_active_dishes"
	allStoppedDishes          = "all_stopped_dishes"
	createOrganizationCommand = "create_organization"
)

type Admin struct {
	bot               *tgbotapi.BotAPI
	updatesChan       chan tgbotapi.Update
	org               service.Organization
	menu              service.Menu
	msgStore          *storage.Messages
	adminID           int64
	startedLunchTime  time.Duration
	finishedLunchTime time.Duration

	// if true - state to activate dishes
	state bool
}

func NewAdmin(bot *tgbotapi.BotAPI, updatesChan chan tgbotapi.Update, org service.Organization, menu service.Menu,
	msgStore *storage.Messages, adminID int64, startedLunchTime time.Duration, finishedLunchTime time.Duration) *Admin {
	return &Admin{
		bot:               bot,
		updatesChan:       updatesChan,
		org:               org,
		menu:              menu,
		msgStore:          msgStore,
		adminID:           adminID,
		startedLunchTime:  startedLunchTime,
		finishedLunchTime: finishedLunchTime,
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

				case createOrganizationCommand:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, createOrganization)
					_, err = a.bot.Send(msg)
					if err != nil {
						logrus.Errorf("createOrganization: send: %s", err.Error())
						continue
					}

					a.msgStore.WaitMessage(update.SentFrom().ID, storage.CreateOrganization, update.Message.MessageID+2)
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

				msgType, ok := a.msgStore.Extract(update.SentFrom().ID)
				if !ok {
					continue
				}
				switch msgType.Action {
				case storage.CreateOrganization:
					err = a.createOrganization(ctx, update.SentFrom().ID, update.Message.Chat.ID, update.Message.Text, update.Message.MessageID)
					if err != nil {
						logrus.Errorf("createOrganization: %s", err.Error())
						continue
					}
					continue

				case storage.AddAddress:
					err = a.addAddress(ctx, update.SentFrom().ID, update.Message.Chat.ID, update.Message.Text)
					if err != nil {
						logrus.Errorf("addAddress: %s", err.Error())
						continue
					}
					continue
				}
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

func (a *Admin) createOrganization(ctx context.Context, userTelegramID, chatID int64, message string, messageID int) error {
	// format message: create Название организации 12:30
	// 12:30 - lunchTime
	if len(strings.Split(message, " ")) < 2 {
		msg := tgbotapi.NewMessage(chatID, "Вы ввели некорректную строку. Попробуйте ещё раз")
		_, err := a.bot.Send(msg)
		if err != nil {
			return fmt.Errorf("send: %w", err)
		}
		a.msgStore.WaitMessage(userTelegramID, storage.CreateOrganization, messageID+2)
		return nil
	}
	organization, errHandle := a.handleCreateOrganization(message)
	if errHandle != "" {
		msg := tgbotapi.NewMessage(chatID, errHandle)
		_, errSend := a.bot.Send(msg)
		if errSend != nil {
			return fmt.Errorf("send: %w", errSend)
		}
		a.msgStore.WaitMessage(userTelegramID, storage.CreateOrganization, messageID+2)
		return nil
	}

	newCtx, cancel := context.WithTimeout(ctx, time.Minute)
	err := a.org.Add(newCtx, organization)
	if err != nil {
		cancel()
		return fmt.Errorf("add: %w", err)
	}
	cancel()

	a.msgStore.WaitMessage(userTelegramID, storage.AddAddress, messageID+2)

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(successfulOrganizationRegistered, organization.Name))
	_, err = a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	msg = tgbotapi.NewMessage(chatID, organization.ID.String())
	_, err = a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	<-time.After(3 * time.Second)

	msg = tgbotapi.NewMessage(chatID, addAddressAfterCreateOrganizationMessage)
	_, err = a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (a *Admin) handleCreateOrganization(message string) (*model.Organization, string) {
	fields := strings.Fields(message)
	lunchTime := fields[len(fields)-1:]
	logrus.Debugf("handleCreateOrganization: luchTime: %s", lunchTime[0])
	splitLunchTime := strings.Split(lunchTime[0], ":")
	if len(splitLunchTime) != 2 {
		return nil, "Вы ввели некорректно время обеда. Попробуйте ещё раз."
	}
	hours, err := strconv.Atoi(splitLunchTime[0])
	if err != nil {
		return nil, "Вы ввели некорректно время обеда. Попробуйте ещё раз"
	}
	if hours > 23 {
		return nil, "Вы ввели некорректно время обеда. Значение часов не может быть больше 23. Попробуйте ещё раз"
	}
	minutes, err := strconv.Atoi(splitLunchTime[1])
	if err != nil {
		return nil, "Вы ввели некорректно время обеда. Попробуйте ещё раз"
	}
	if minutes > 59 {
		return nil, "Вы ввели некорректно время обеда. Значение минут не может быть больше 59. Попробуйте ещё раз"
	}
	minute := int(a.finishedLunchTime.Minutes()) % 60
	if hours > int(a.finishedLunchTime.Hours()) || hours == int(a.finishedLunchTime.Hours()) && minutes > minute {
		return nil, fmt.Sprintf(tooLateLunchTimeMessage, int(a.finishedLunchTime.Hours()), minute)
	}
	minute = int(a.startedLunchTime.Minutes()) % 60
	if hours < int(a.startedLunchTime.Hours()) || hours == int(a.startedLunchTime.Hours()) && minutes < minute {
		return nil, fmt.Sprintf(tooEarlyLunchTimeMessage, int(a.startedLunchTime.Hours()), minute)
	}
	logrus.Debugf("handleCreateOrganization: hours: %d, minutes: %d", hours, minutes)
	orgName := strings.Join(fields[:len(fields)-1], " ")
	logrus.Debugf("handleCreateOrganization: orgName: %s", orgName)
	return &model.Organization{
		ID:        uuid.New(),
		Name:      orgName,
		LunchTime: time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute,
	}, ""
}

func (a *Admin) addAddress(ctx context.Context, userTelegramID, chatID int64, message string) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Minute)
	err := a.org.UpdateAddress(newCtx, userTelegramID, message)
	if err != nil {
		cancel()
		return fmt.Errorf("updateAddress: %w", err)
	}
	cancel()

	msg := tgbotapi.NewMessage(chatID, successfulAddAddress)
	_, err = a.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}
