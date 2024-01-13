package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/producer"
	"github.com/chucky-1/food-delivery-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/chucky-1/food-delivery-bot/internal/config"
	"github.com/chucky-1/food-delivery-bot/internal/consumer"
	"github.com/chucky-1/food-delivery-bot/internal/repository"
	"github.com/chucky-1/food-delivery-bot/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	logrus.SetLevel(logrus.Level(cfg.LogLevel))

	pool, err := pgxpool.Connect(ctx, cfg.Endpoint)
	if err != nil {
		logrus.Fatalf("couldn't connect to database: %v", err)
	}
	if err = pool.Ping(ctx); err != nil {
		logrus.Fatalf("couldn't ping database: %v", err)
	}

	dishesByCategories, allDishes := parseMenu(&cfg.Menu)

	transactorRep := repository.NewTransactor(pool)
	userRep := repository.NewUser(transactorRep)
	orgRep := repository.NewOrganization(transactorRep)
	telegramUserRep := repository.NewTelegram(transactorRep)
	orderRep := repository.NewOrder(transactorRep, cfg.Timezone, cfg.PeriodOfTimeBeforeLunchToShipOrder)
	menuRep := repository.NewMenu(cfg.Menu.Categories, dishesByCategories, dishesByCategories, make(map[string][]*model.Dish), allDishes)

	authService := service.NewAuth(userRep, telegramUserRep, orgRep, transactorRep)
	orgService := service.NewOrganization(orgRep)
	menuService := service.NewMenu(menuRep)
	orderService := service.NewOrder(orderRep)
	telegramService := service.NewTelegram(telegramUserRep)
	statisticsService := service.NewStatistics(orderRep, transactorRep)

	msgStore := storage.NewMessage()

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBot.Token)
	if err != nil {
		logrus.Fatal(err)
	}
	//bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.TelegramBot.Timeout
	updatesChan := bot.GetUpdatesChan(u)

	adminChan := make(chan tgbotapi.Update)
	botConsumer := consumer.NewBot(bot, updatesChan, authService, orgService, menuService, orderService, msgStore, cfg.Timezone,
		cfg.StartedLunchTime, cfg.FinishedLunchTime, cfg.AdminChatID, adminChan)
	go botConsumer.Consume(ctx)

	adminConsumer := consumer.NewAdmin(bot, adminChan, menuService, cfg.AdminChatID)
	go adminConsumer.Consume(ctx)

	usersReminder := producer.NewUsersReminder(bot, telegramService, orderService, cfg.Timezone, cfg.StartingMinutes, cfg.TickInterval,
		cfg.PeriodOfTimeBeforeLunchToShipOrder, cfg.FirstReminder, cfg.SecondReminder)
	go usersReminder.Remind(ctx)

	orderSender := producer.NewOrderSender(bot, orderService, cfg.Timezone, cfg.StartingMinutes, cfg.TickInterval,
		cfg.PeriodOfTimeBeforeLunchToShipOrder, cfg.AdminChatID)
	go orderSender.Send(ctx)

	statisticsSender := producer.NewStatisticsSender(bot, statisticsService, cfg.Timezone, cfg.ReportHour, cfg.ReportReceivers)
	go statisticsSender.StatisticsSend(ctx)

	// http server to check health
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err = io.WriteString(writer, "")
		if err != nil {
			logrus.Fatalf("couldn't write response: %v", err)
		}
	})
	go func() {
		err = http.ListenAndServe(":8080", nil)
		if err != nil {
			logrus.Fatalf("couldn't listen and serve: %v", err)
		}
	}()

	logrus.Infof("app has started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)
	<-quit
	cancel()
	<-time.After(2 * time.Second)
}

func parseMenu(menu *config.Menu) (map[string][]*model.Dish, map[string]*model.Dish) {
	dishesByCategories := make(map[string][]*model.Dish)
	allDishes := make(map[string]*model.Dish)
	for _, category := range menu.Categories {
		switch category {
		case model.Soups:
			for name, price := range menu.Soups {
				dish := model.Dish{
					Name:     name,
					Price:    price,
					Category: model.Soups,
				}
				dishesByCategories[model.Soups] = append(dishesByCategories[model.Soups], &dish)
				allDishes[dish.String()] = &dish
			}
		case model.Salads:
			for name, price := range menu.Salads {
				dish := model.Dish{
					Name:     name,
					Price:    price,
					Category: model.Salads,
				}
				dishesByCategories[model.Salads] = append(dishesByCategories[model.Salads], &dish)
				allDishes[dish.String()] = &dish
			}
		case model.MainCourse:
			for name, price := range menu.MainCourse {
				dish := model.Dish{
					Name:     name,
					Price:    price,
					Category: model.MainCourse,
				}
				dishesByCategories[model.MainCourse] = append(dishesByCategories[model.MainCourse], &dish)
				allDishes[dish.String()] = &dish
			}
		case model.Desserts:
			for name, price := range menu.Desserts {
				dish := model.Dish{
					Name:     name,
					Price:    price,
					Category: model.Desserts,
				}
				dishesByCategories[model.Desserts] = append(dishesByCategories[model.Desserts], &dish)
				allDishes[dish.String()] = &dish
			}
		case model.Drinks:
			for name, price := range menu.Drinks {
				dish := model.Dish{
					Name:     name,
					Price:    price,
					Category: model.Drinks,
				}
				dishesByCategories[model.Drinks] = append(dishesByCategories[model.Drinks], &dish)
				allDishes[dish.String()] = &dish
			}
		}
	}
	return dishesByCategories, allDishes
}
