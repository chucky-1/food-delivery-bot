package consumer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	start    = "start"
	register = "register"

	createOrganization = "create"
	joinOrganization   = "join"
	addAddress         = "address"

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
	successfulRegistered = "🎉 Поздравляем вас с успешной регистрацией! 🎉\n\n" +
		"Теперь вы можете создать свою организацию или присоединиться к существующей.\n\n" +
		"Для создания новой организации отправьте нам сообщение в следующем формате:\n\n" +
		"create Название организации 12:30\n\n" +
		"Где 12:30 - это время, к которому вы хотели бы получить свой обед.\n\n" +
		"Если вы хотите присоединиться к уже существующей организации, отправьте сообщение в следующем формате:\n\n" +
		"join 59efbd76\n\n" +
		"Где 59efbd76 - это уникальный идентификатор вашей организации. Вы можете получить этот идентификатор у пользователя, который создал организацию."
	successfulOrganizationRegistered = "🎉 Поздравляем! Вы успешно зарегистрировали свою организацию: %s\n\n" +
		"Теперь, чтобы присоединиться к ней, потребуется уникальный идентификатор (ID), который мы пришлем в следующем сообщении.\n\n" +
		"Вам не нужно вступать в свою организауцию. Вы уже состоите в ней."
	successfulJoinOrganization   = "🎉 Поздравляем! Вы успешно вступили в организацию! 🎉"
	successfulClearOrder         = "😊 Мы удалили всё из вашего заказа"
	successfulConfirmOrder       = "🎉 Заказ успешно подтверждён! Он будет передан нашему администратору вместе с другими заказами для вашей организации. Спасибо за выбор нас! Приятного аппетита! 😊"
	successfulCancelOrder        = "😊 Вы успешно отменили заказ"
	userAlreadyHasConfirmedOrder = "В данный момент, изменение вашего заказа недоступно, однако вы можете его отменить и создать новый заказ, если необходимо."
	addAddressMessage            = "🏢 Теперь давайте добавим адрес вашей организации, чтобы мы знали, куда доставлять ваши обеды. Просто отправьте сообщение с адресом в следующем формате:\n\n" +
		"Address ул. Толбухина 18/2\n\n" +
		"Вы можете предоставить адрес в любой форме и даже добавить комментарии. Главное, чтобы мы точно знали, куда направлять ваш заказ. " +
		"Если в будущем адрес изменится, не забудьте обновить его, отправив нам подобное сообщение с новыми данными."
	successfulAddAddress = "🎉 Вы успешно добавили адрес организации!"
	menuRequest          = "📋 Чтобы посмотреть наше меню, отправьте команду /menu или просто напишите \"Меню\". Так вы сможете ознакомиться с нашим разнообразным выбором блюд и выбрать то, что подходит именно вам!"
)

type Bot struct {
	bot         *tgbotapi.BotAPI
	updatesChan tgbotapi.UpdatesChannel
	auth        service.Auth
	org         service.Organization
	menu        service.Menu
	order       service.Order
	timezone    time.Duration
}

func NewBot(bot *tgbotapi.BotAPI, updatesChan tgbotapi.UpdatesChannel, auth service.Auth, org service.Organization,
	menu service.Menu, order service.Order, timezone time.Duration) *Bot {
	return &Bot{
		bot:         bot,
		updatesChan: updatesChan,
		auth:        auth,
		org:         org,
		menu:        menu,
		order:       order,
		timezone:    timezone,
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
						ID:        update.SentFrom().ID,
						ChatID:    update.Message.Chat.ID,
						FirstName: update.SentFrom().FirstName,
					})
					if err != nil {
						logrus.Errorf("registerCommand: %s", err.Error())
						cancel()
						continue
					}
					cancel()
					logrus.Debugf("user registered: %s %d", update.SentFrom().UserName, update.SentFrom().ID)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, successfulRegistered)
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Error("register send: %s", err.Error())
						continue
					}
					continue
				case menu:
					err := b.sendMenu(ctx, update.SentFrom().ID, update.Message.Chat.ID)
					if err != nil {
						logrus.Errorf("sendMenu: %s", err.Error())
						continue
					}
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
					err := b.order.ClearOrdersByUser(newCtx, update.SentFrom().ID, time.Now().UTC().Add(b.timezone))
					if err != nil {
						cancel()
						logrus.Error("cancelOrder: %w", err)
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

				logrus.Debugf("consume: message: %s", update.Message.Text)
				split := strings.Split(update.Message.Text, " ")
				firstWord := strings.ToLower(split[0])
				logrus.Debugf("consume: firstWorld: %s", firstWord)
				messageWithoutFirstWord := strings.Join(split[len(split)-(len(split)-1):], " ")
				logrus.Debugf("consume: messageWithoutFirstWord: %s", messageWithoutFirstWord)
				switch firstWord {
				case createOrganization:
					// format message: create Название организации 12:30
					// 12:30 - lunchTime
					if len(strings.Split(messageWithoutFirstWord, " ")) < 2 {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы ввели некорректную строку")
						_, err = b.bot.Send(msg)
						if err != nil {
							logrus.Errorf("createOrganization send: %s", err.Error())
							continue
						}
						continue
					}
					organization, userError := handleCreateOrganization(messageWithoutFirstWord)
					if userError != "" {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, userError)
						_, err = b.bot.Send(msg)
						if err != nil {
							logrus.Errorf("createOrganization send: %s", err.Error())
							continue
						}
						continue
					}

					userTelegramID := update.SentFrom().ID

					newCtx, cancel = context.WithTimeout(ctx, time.Minute)
					err = b.org.Add(newCtx, organization, userTelegramID)
					if err != nil {
						logrus.Errorf("createOrganization: %s", err.Error())
						cancel()
						continue
					}
					cancel()

					newCtx, cancel = context.WithTimeout(ctx, time.Minute)
					if err = b.org.Join(newCtx, organization.ID, userTelegramID); err != nil {
						logrus.Errorf("createOrganization: %s", err.Error())
						cancel()
						continue
					}
					cancel()

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(successfulOrganizationRegistered, organization.Name))
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Errorf("createOrganization: send: %s", err.Error())
						continue
					}

					msg = tgbotapi.NewMessage(update.Message.Chat.ID, organization.ID.String())
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Errorf("createOrganization: send: %s", err.Error())
						continue
					}

					msg = tgbotapi.NewMessage(update.Message.Chat.ID, addAddressMessage)
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Errorf("createOrganization: send: %s", err.Error())
						continue
					}

					continue
				case joinOrganization:
					uid, errParse := uuid.Parse(messageWithoutFirstWord)
					if errParse != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вы ввели некорректную строку")
						_, errSend := b.bot.Send(msg)
						if errSend != nil {
							logrus.Errorf("joinOrganization send: %s", errSend.Error())
							continue
						}
						continue
					}

					newCtx, cancel = context.WithTimeout(ctx, time.Minute)
					if err = b.org.Join(newCtx, uid, update.SentFrom().ID); err != nil {
						logrus.Errorf("joinOrganization: %s", err.Error())
						cancel()
						continue
					}
					cancel()

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, successfulJoinOrganization)
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Errorf("joinOrganization send: %s", err.Error())
						continue
					}

					msg = tgbotapi.NewMessage(update.Message.Chat.ID, menuRequest)
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Errorf("joinOrganization send: %s", err.Error())
						continue
					}

					continue
				case addAddress:
					newCtx, cancel = context.WithTimeout(ctx, time.Minute)
					err = b.org.UpdateAddress(newCtx, update.Message.Chat.ID, messageWithoutFirstWord)
					if err != nil {
						logrus.Errorf("addAddress: %s", err.Error())
						cancel()
						continue
					}
					cancel()

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, successfulAddAddress)
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Errorf("addAddress: %s", err.Error())
						continue
					}

					msg = tgbotapi.NewMessage(update.Message.Chat.ID, menuRequest)
					_, err = b.bot.Send(msg)
					if err != nil {
						logrus.Errorf("joinOrganization send: %s", err.Error())
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
	dishes, err := b.menu.GetAllDishesByCategory(newCtx, category)
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
	message = fmt.Sprintf("%s\nСумма вашего заказа: %.2f", message, totalPrice)
	msg := tgbotapi.NewMessage(chatID, message)
	_, err = b.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func handleCreateOrganization(message string) (*model.Organization, string) {
	fields := strings.Fields(message)
	lunchTime := fields[len(fields)-1:]
	logrus.Debugf("handleCreateOrganization: luchTime: %s", lunchTime[0])
	splitLunchTime := strings.Split(lunchTime[0], ":")
	if len(splitLunchTime) != 2 {
		return nil, "Вы ввели некорректно время обеда"
	}
	hours, err := strconv.Atoi(splitLunchTime[0])
	if err != nil {
		return nil, "Вы ввели некорректно время обеда"
	}
	if hours > 23 {
		return nil, "Вы ввели некорректно время обеда. Значение часов не может быть больше 23"
	}
	minutes, err := strconv.Atoi(splitLunchTime[1])
	if err != nil {
		return nil, "Вы ввели некорректно время обеда"
	}
	if minutes > 59 {
		return nil, "Вы ввели некорректно время обеда. Значение минут не может быть больше 59"
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
