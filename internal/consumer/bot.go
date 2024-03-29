package consumer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
	"github.com/chucky-1/food-delivery-bot/internal/service"
	"github.com/chucky-1/food-delivery-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	start    = "start"
	register = "register"

	menu         = "menu"
	goBackToMenu = "Вернуться в меню"
	confirmOrder = "Подтвердить заказ"
	clearOrder   = "Очистить заказ"
	cancelOrder  = "Отменить заказ"
)

var (
	welcomeMessage = "🍽 Добро пожаловать в бот кафе «Крипта»! 🍽\n\n" +
		"Мы рады приветствовать Вас в нашем уютном кафе, где вы можете насладиться вкусными обедами, не покидая здание своего офиса.\n\n" +
		"Что мы предлагаем:\n" +
		"🥗 Разнообразное меню обедов на любой вкус – от классических блюд до эксклюзивных гастрономических изысков.\n" +
		"🚀 Быстрая и надежная доставка прямо к вам, чтобы вы могли наслаждаться своим обедом в комфорте.\n" +
		"🌟 Качество и свежесть ингредиентов – мы заботимся о вашем здоровье и удовольствии от еды.\n" +
		"📋 Удобный заказ через этого бота – всего несколько кликов, и ваш обед уже в пути!\n\n" +
		"Не забудьте посмотреть наше меню и сделать свой первый заказ. Мы уверены, что вы останетесь довольны!\n\n" +
		"Если у вас есть какие-либо вопросы или пожелания, не стесняйтесь обращаться к нам. Мы всегда готовы сделать ваш обед особенным.\n\n" +
		"Приятного аппетита! 🍽😊\n\n" +
		"/register"
	startRegister        = "Ведите ваше имя"
	inputLastName        = "Введите фамилию"
	inputMiddleName      = "Введите отчество"
	successfulRegistered = "🎉 Поздравляем вас с успешной регистрацией! 🎉\n\n" +
		"Для вступления в организацию нажмите /join"
	joinToOrganization           = "Введите ID организации \n\n"
	successfulJoinOrganization   = "🎉 Поздравляем! Вы успешно вступили в организацию! 🎉"
	successfulClearOrder         = "😊 Мы удалили всё из вашего заказа"
	successfulConfirmOrder       = "🎉 Заказ успешно подтверждён! Он будет передан нашему администратору вместе с другими заказами для вашей организации. Спасибо за выбор нас! Приятного аппетита! 😊"
	successfulCancelOrder        = "😊 Вы успешно отменили заказ"
	userAlreadyHasConfirmedOrder = "В данный момент, изменение вашего заказа недоступно, однако вы можете его отменить и создать новый заказ, если необходимо."
	menuRequest                  = "📋 Чтобы посмотреть наше меню, отправьте команду /menu или просто напишите \"Меню\". Так вы сможете ознакомиться с нашим разнообразным выбором блюд и выбрать то, что подходит именно вам!"
	lunchTimePassed              = "Извините, но время обеда уже прошло или заказы вашей организации уже отправлены. Обратитесь к администратору за помощью @kriptabar"
	cannotCancelOrderMessage     = "Извините, но мы не можем отменить ваш заказ. Он уже отправлен администратору. " +
		"Если вы хотите это сделать, свяжитесь с нами @kriptabar"
	tooLateLunchTimeMessage  = "Вы ввели слишком поздее время обеда. Самое поздее возможное время обеда: %d:%d. Попробуйте ещё раз."
	tooEarlyLunchTimeMessage = "Вы ввели слишком раннее время обеда. Мы начинаем доставлять обеды с %d:%d. Попробуйте ещё раз."
	weekendMessage           = "Извините, но сегодня выходной ☺"
	errJoinToOrganization    = "Что то пошло не так, скорее всего такой организации не существует, проверьте ID"
)

type Bot struct {
	bot         *tgbotapi.BotAPI
	updatesChan tgbotapi.UpdatesChannel
	auth        service.Auth
	org         service.Organization
	menu        service.Menu
	order       service.Order
	msgStore    *storage.Messages
	timezone    time.Duration
	adminID     int64
	adminChan   chan tgbotapi.Update
}

func NewBot(bot *tgbotapi.BotAPI, updatesChan tgbotapi.UpdatesChannel, auth service.Auth, org service.Organization,
	menu service.Menu, order service.Order, msgStore *storage.Messages, timezone time.Duration, adminID int64,
	adminChan chan tgbotapi.Update) *Bot {
	return &Bot{
		bot:         bot,
		updatesChan: updatesChan,
		auth:        auth,
		org:         org,
		menu:        menu,
		order:       order,
		msgStore:    msgStore,
		timezone:    timezone,
		adminID:     adminID,
		adminChan:   adminChan,
	}
}

func (b *Bot) Consume(ctx context.Context) {
	logrus.Info("bot consumer started")
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("bot consumer stopped: %s", ctx.Err().Error())
			return
		case update := <-b.updatesChan:
			if update.SentFrom().ID == b.adminID {
				b.adminChan <- update
				continue
			}
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case start:
					logrus.Debugf("start: %s %d", update.SentFrom().UserName, update.SentFrom().ID)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMessage)
					_, err := b.bot.Send(msg)
					if err != nil {
						logrus.Error("start send: %s", err.Error())
						continue
					}
					continue
				case register:
					newCtx, cancel := context.WithTimeout(ctx, time.Minute)
					err := b.auth.Register(newCtx, &model.TelegramUser{
						ID:       update.SentFrom().ID,
						ChatID:   update.Message.Chat.ID,
						Username: update.SentFrom().UserName,
					})
					if err != nil {
						logrus.Errorf("registerCommand: %s", err.Error())
						cancel()
						continue
					}
					cancel()

					logrus.Debugf("user registered: %s %d", update.SentFrom().UserName, update.SentFrom().ID)

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, startRegister)
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Error("register send: %s", err.Error())
						continue
					}
					b.msgStore.WaitMessage(update.SentFrom().ID, storage.AddFirstName, update.Message.MessageID+2, "")
					continue
				case menu:
					err := b.sendMenu(ctx, update.SentFrom().ID, update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("sendMenu: %s", err.Error())
						continue
					}
					continue

				case storage.JoinToOrganization:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, joinToOrganization)
					_, err := b.bot.Send(msg)
					if err != nil {
						logrus.Errorf("createOrganization: send: %s", err.Error())
						continue
					}

					b.msgStore.WaitMessage(update.SentFrom().ID, storage.JoinToOrganization, update.Message.MessageID+2, "")
					continue
				}
			} else {
				switch update.Message.Text {
				case model.Soups:
					err := b.sendDishes(ctx, update.SentFrom().ID, model.Soups, update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case model.Salads:
					err := b.sendDishes(ctx, update.SentFrom().ID, model.Salads, update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case model.MainCourse:
					err := b.sendDishes(ctx, update.SentFrom().ID, model.MainCourse, update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case model.Desserts:
					err := b.sendDishes(ctx, update.SentFrom().ID, model.Desserts, update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case model.Drinks:
					err := b.sendDishes(ctx, update.SentFrom().ID, model.Drinks, update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("sendDishes: %s", err.Error())
						continue
					}
					continue
				case goBackToMenu, "Меню":
					err := b.sendMenu(ctx, update.SentFrom().ID, update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("sendMenu: %s", err.Error())
						continue
					}
					continue
				case clearOrder:
					newCtx, cancel := context.WithTimeout(ctx, time.Minute)
					err := b.order.ClearOrdersByUser(newCtx, update.SentFrom().ID, time.Now().UTC().Add(b.timezone))
					if err != nil {
						cancel()
						logrus.Errorf("clearOrders: %s", err.Error())
						continue
					}

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, successfulClearOrder)
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					_, err = b.bot.Send(msg)
					if err != nil {
						cancel()
						logrus.Errorf("clearOrder: send: %s", err.Error())
					}

					err = b.sendMenu(newCtx, update.SentFrom().ID, update.Message.Chat.ID)
					if err != nil {
						cancel()
						logrus.Errorf("clearOrder: %s", err.Error())
					}
					cancel()
					continue
				case confirmOrder:
					newCtx, cancel := context.WithTimeout(ctx, time.Minute)
					err := b.order.ConfirmOrderByUser(newCtx, update.SentFrom().ID)
					if err != nil {
						logrus.Error(err.Error())
						cancel()
						continue
					}
					cancel()

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, successfulConfirmOrder)
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Errorf("confirmOrder: send: %s", err.Error())
						continue
					}
					continue
				case cancelOrder:
					newCtx, cancel := context.WithTimeout(ctx, time.Minute)
					err := b.order.ClearOrdersByUserWithCheckLunchTime(newCtx, update.SentFrom().ID, time.Now().UTC().Add(b.timezone))
					if err != nil {
						cancel()
						switch {
						case errors.Is(err, repository.ErrLunchTimePassed):
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, cannotCancelOrderMessage)
							_, errSend := b.bot.Send(msg)
							if errSend != nil {
								logrus.Errorf("cancelOrder: send: %s", errSend.Error())
								continue
							}
							continue
						}
						logrus.Error("cancelOrder: %s", err.Error())
						continue
					}

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, successfulCancelOrder)
					_, err = b.bot.Send(msg)
					if err != nil {
						cancel()
						logrus.Errorf("cancelOrders: send: %s", err.Error())
					}

					err = b.sendMenu(newCtx, update.SentFrom().ID, update.Message.Chat.ID)
					if err != nil {
						cancel()
						logrus.Errorf("cancelOrder: %s", err.Error())
					}
					cancel()
					continue
				}

				newCtx, cancel := context.WithTimeout(ctx, time.Minute)
				dish, err := b.menu.GetDish(newCtx, update.Message.Text)
				if err != nil {
					logrus.Error(err.Error())
					cancel()
					continue
				}
				cancel()
				if dish != nil {
					err = b.addDishInOrder(ctx, dish, update.SentFrom().ID, update.Message.Chat.ID)
					if err != nil {
						switch {
						case errors.Is(err, service.ErrWeekend):
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, weekendMessage)
							_, errSend := b.bot.Send(msg)
							if errSend != nil {
								logrus.Errorf("addDishInOrder: send: %s", errSend.Error())
								continue
							}
							continue
						case errors.Is(err, repository.ErrLunchTimePassed):
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, lunchTimePassed)
							_, errSend := b.bot.Send(msg)
							if errSend != nil {
								logrus.Errorf("addDishInOrder: send: %s", errSend.Error())
								continue
							}
							continue
						}
						logrus.Error(err.Error())
						continue
					}

					newCtx, cancel = context.WithTimeout(ctx, time.Minute)
					err = b.sendDishes(newCtx, update.SentFrom().ID, dish.Category, update.Message.Chat.ID)
					if err != nil {
						logrus.Error(err.Error())
						cancel()
						continue
					}
					cancel()
					continue
				}

				msgType, ok := b.msgStore.Extract(update.SentFrom().ID)
				if !ok {
					continue
				}
				switch msgType.Action {
				case storage.JoinToOrganization:
					err = b.joinToOrganization(ctx, update.SentFrom().ID, update.Message.Chat.ID, update.Message.Text, update.Message.MessageID)
					if err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, errJoinToOrganization)
						_, errSend := b.bot.Send(msg)
						if errSend != nil {
							logrus.Errorf("joinToOrganization: send: %s", err.Error())
							continue
						}
						b.msgStore.WaitMessage(update.SentFrom().ID, storage.JoinToOrganization, update.Message.MessageID+2, "")
						logrus.Errorf("joinToOrganization: %s", err.Error())
						continue
					}
					continue
				case storage.AddFirstName:
					err = b.auth.UpdateFirstName(ctx, int(update.SentFrom().ID), update.Message.Text)
					if err != nil {
						logrus.Errorf("addFirstName: user_telegram_id: %d, first_name: %s, err: %s",
							int(update.SentFrom().ID), update.Message.Text, err.Error())
						continue
					}

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, inputLastName)
					_, errSend := b.bot.Send(msg)
					if errSend != nil {
						logrus.Errorf("addFirstName: send: %s", err.Error())
						continue
					}

					b.msgStore.WaitMessage(update.SentFrom().ID, storage.AddLastName, update.Message.MessageID+2, "")

				case storage.AddLastName:
					err = b.auth.UpdateLastName(ctx, int(update.SentFrom().ID), update.Message.Text)
					if err != nil {
						logrus.Errorf("addLastName: user_telegram_id: %d, last_name: %s, err: %s",
							int(update.SentFrom().ID), update.Message.Text, err.Error())
						continue
					}

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, inputMiddleName)
					_, errSend := b.bot.Send(msg)
					if errSend != nil {
						logrus.Errorf("addLastName: send: %s", err.Error())
						continue
					}

					b.msgStore.WaitMessage(update.SentFrom().ID, storage.AddMiddleName, update.Message.MessageID+2, "")

				case storage.AddMiddleName:
					err = b.auth.UpdateMiddleName(ctx, int(update.SentFrom().ID), update.Message.Text)
					if err != nil {
						logrus.Errorf("updateMiddleName: user_telegram_id: %d, last_name: %s, err: %s",
							int(update.SentFrom().ID), update.Message.Text, err.Error())
						continue
					}

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, successfulRegistered)
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Error("register send: %s", err.Error())
						continue
					}
					continue
				}
			}
		}
	}
}

func (b *Bot) sendMenu(ctx context.Context, userTelegramID, chatID int64) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Minute)
	isUserHaveConfirmedOrder, err := b.order.IsUserHaveConfirmedOrder(newCtx, userTelegramID)
	if err != nil {
		cancel()
		return err
	}
	if isUserHaveConfirmedOrder {
		msg := tgbotapi.NewMessage(chatID, userAlreadyHasConfirmedOrder)
		var buttons [][]tgbotapi.KeyboardButton
		but := tgbotapi.NewKeyboardButton(cancelOrder)
		row := tgbotapi.NewKeyboardButtonRow(but)
		buttons = append(buttons, row)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons...)
		_, err = b.bot.Send(msg)
		if err != nil {
			cancel()
			return fmt.Errorf("send: %w", err)
		}
		cancel()
		return nil
	}

	categories, err := b.menu.GetAllCategories(newCtx)
	if err != nil {
		cancel()
		return err
	}

	msg := tgbotapi.NewMessage(chatID, "Меню")
	var buttons [][]tgbotapi.KeyboardButton
	for _, category := range categories {
		but := tgbotapi.NewKeyboardButton(category)
		row := tgbotapi.NewKeyboardButtonRow(but)
		buttons = append(buttons, row)
	}

	exist, err := b.order.IsUserHaveAnyOrders(newCtx, userTelegramID)
	if err != nil {
		cancel()
		return err
	}
	cancel()
	if exist {
		but := tgbotapi.NewKeyboardButton(confirmOrder)
		row := tgbotapi.NewKeyboardButtonRow(but)
		buttons = append(buttons, row)

		but = tgbotapi.NewKeyboardButton(clearOrder)
		row = tgbotapi.NewKeyboardButtonRow(but)
		buttons = append(buttons, row)
	}

	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons...)
	_, err = b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (b *Bot) sendDishes(ctx context.Context, userTelegramID int64, category string, chatID int64) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Minute)
	dishes, err := b.menu.GetActiveDishesByCategory(newCtx, category)
	if err != nil {
		cancel()
		return err
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

	exist, err := b.order.IsUserHaveAnyOrders(newCtx, userTelegramID)
	if err != nil {
		cancel()
		return err
	}
	cancel()
	if exist {
		but = tgbotapi.NewKeyboardButton(confirmOrder)
		row = tgbotapi.NewKeyboardButtonRow(but)
		buttons = append(buttons, row)

		but = tgbotapi.NewKeyboardButton(clearOrder)
		row = tgbotapi.NewKeyboardButtonRow(but)
		buttons = append(buttons, row)
	}

	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons...)
	_, err = b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (b *Bot) addDishInOrder(ctx context.Context, dish *model.Dish, userTelegramID int64, chatID int64) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Minute)
	err := b.order.AddDish(newCtx, dish, userTelegramID)
	if err != nil {
		cancel()
		return fmt.Errorf("addDishInOrder: %w", err)
	}

	dishesByCategories, err := b.order.GetAllDishesByCategory(newCtx, userTelegramID)
	if err != nil {
		cancel()
		return fmt.Errorf("addDishInOrder: %w", err)
	}
	categories, err := b.menu.GetAllCategories(newCtx)
	if err != nil {
		cancel()
		return fmt.Errorf("addDishInOrder: %w", err)
	}
	cancel()

	var (
		message    = "Ваш заказ:\n\n"
		totalPrice float32
	)
	for _, category := range categories {
		dishes, ok := dishesByCategories[category]
		if !ok {
			continue
		}
		for _, d := range dishes {
			message = fmt.Sprintf("%s%s\n", message, d.Name)
			totalPrice += d.Price
		}
	}
	message = fmt.Sprintf("%s\nСумма вашего заказа: %.2f\n\nЧто бы отправить заказ, нажмите «Подтвердить заказ»", message, totalPrice)
	msg := tgbotapi.NewMessage(chatID, message)
	_, err = b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (b *Bot) joinToOrganization(ctx context.Context, userTelegramID, chatID int64, message string, messageID int) error {
	uid, errParse := uuid.Parse(message)
	if errParse != nil {
		msg := tgbotapi.NewMessage(chatID, "Вы ввели некорректную строку. Попробуйте ещё раз")
		_, err := b.bot.Send(msg)
		if err != nil {
			return fmt.Errorf("send: %w", err)
		}
		b.msgStore.WaitMessage(userTelegramID, storage.JoinToOrganization, messageID+2, "")
		return nil
	}

	newCtx, cancel := context.WithTimeout(ctx, time.Minute)
	if err := b.org.Join(newCtx, uid, userTelegramID); err != nil {
		cancel()
		return fmt.Errorf("join: %w", err)
	}
	cancel()

	msg := tgbotapi.NewMessage(chatID, successfulJoinOrganization)
	_, err := b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	msg = tgbotapi.NewMessage(chatID, menuRequest)
	_, err = b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}
