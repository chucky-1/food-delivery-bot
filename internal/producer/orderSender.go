package producer

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type OrderSender struct {
	bot                                *tgbotapi.BotAPI
	order                              service.Order
	timezone                           time.Duration
	startingMinutes                    []int
	tickInterval                       time.Duration
	periodOfTimeBeforeLunchToShipOrder time.Duration
	adminChatID                        int64
}

func NewOrderSender(bot *tgbotapi.BotAPI, order service.Order, timezone time.Duration, startingMinutes []int, tickInterval time.Duration,
	periodOfTimeBeforeLunchToShipOrder time.Duration, adminChatID int64) *OrderSender {
	return &OrderSender{
		bot:                                bot,
		order:                              order,
		timezone:                           timezone,
		startingMinutes:                    startingMinutes,
		tickInterval:                       tickInterval,
		periodOfTimeBeforeLunchToShipOrder: periodOfTimeBeforeLunchToShipOrder,
		adminChatID:                        adminChatID,
	}
}

func (s *OrderSender) Send(ctx context.Context) {
	logrus.Info("orderSender producer started")
	waitTimeToCreateTicker(ctx, s.startingMinutes)
	logrus.Infof("orderSender is ready to create ticker: %s", time.Now().UTC().String())
	t := time.NewTicker(s.tickInterval)
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("orderSender producer stopped: %s", ctx.Err().Error())
			t.Stop()
			return
		case <-t.C:
			truncatedNowWithTimezone := time.Now().UTC().Add(s.timezone).Truncate(time.Minute)
			hour := time.Duration(truncatedNowWithTimezone.Add(s.periodOfTimeBeforeLunchToShipOrder).Hour())
			minute := time.Duration(truncatedNowWithTimezone.Add(s.periodOfTimeBeforeLunchToShipOrder).Minute())
			lunchTime := hour*time.Hour + minute*time.Minute
			logrus.Debugf("orderSender: lunch time: %s", lunchTime.String())

			newCtx, cancel := context.WithTimeout(ctx, time.Minute)
			dataByOrganizationID, err := s.order.GetUserOrdersByOrganizationLunchTime(newCtx, lunchTime.String())
			if err != nil {
				logrus.Errorf("orderSender: %s", err.Error())
				cancel()
				continue
			}
			cancel()
			if len(dataByOrganizationID) == 0 {
				continue
			}

			countOfDishes := make(map[string]int)
			generalMsg := fmt.Sprintf("Заказы к %d:%d\n\n", hour, minute)
			var generalSum float32
			for _, data := range dataByOrganizationID {
				orgMsg := fmt.Sprintf("%s\n%s\n", data.OrganizationName, data.OrganizationAddress)
				var sumByOrg float32
				for _, dishes := range data.DishesByCategories {
					for _, dish := range dishes {
						orgMsg = fmt.Sprintf("%s%s - %d\n", orgMsg, dish.Dish.Name, dish.Count)
						sumByOrg += dish.Dish.Price * float32(dish.Count)
						countOfDishes[dish.Name] += dish.Count
					}
				}
				orgMsg = fmt.Sprintf("%sСумма заказ по организации: %.2f\n\n", orgMsg, sumByOrg)
				generalMsg = fmt.Sprintf("%s%s", generalMsg, orgMsg)
				generalSum += sumByOrg
			}

			tgMsg := tgbotapi.NewMessage(s.adminChatID, generalMsg)
			_, err = s.bot.Send(tgMsg)
			if err != nil {
				logrus.Errorf("orderSender: %s", err.Error())
			}

			msg := fmt.Sprintf("Общий заказ по всем организациям\n")
			for dish, count := range countOfDishes {
				msg = fmt.Sprintf("%s%s - %d\n", msg, dish, count)
			}
			msg = fmt.Sprintf("%sОбщая сумма заказов: %.2f", msg, generalSum)

			tgMsg = tgbotapi.NewMessage(s.adminChatID, msg)
			_, err = s.bot.Send(tgMsg)
			if err != nil {
				logrus.Errorf("orderSender: %s", err.Error())
			}
		}
	}
}
