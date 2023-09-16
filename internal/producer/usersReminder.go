package producer

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

var (
	firstOrderReminderMessage  = "Что бы отправить заказ осталось %.0f минут"
	secondOrderReminderMessage = "Что бы отправить заказ осталось %.0f минут"
)

type UsersReminder struct {
	bot                                *tgbotapi.BotAPI
	telegram                           service.Telegram
	order                              service.Order
	timezone                           time.Duration
	startingMinutes                    []int
	tickInterval                       time.Duration
	periodOfTimeBeforeLunchToShipOrder time.Duration
	firstReminder                      time.Duration
	secondReminder                     time.Duration
}

func NewUsersReminder(bot *tgbotapi.BotAPI, telegram service.Telegram, order service.Order, timezone time.Duration, startingMinutes []int,
	tickInterval time.Duration, periodOfTimeBeforeLunchToShipOrder time.Duration, firstReminder time.Duration, secondReminder time.Duration) *UsersReminder {
	return &UsersReminder{
		bot:                                bot,
		telegram:                           telegram,
		order:                              order,
		timezone:                           timezone,
		startingMinutes:                    startingMinutes,
		tickInterval:                       tickInterval,
		periodOfTimeBeforeLunchToShipOrder: periodOfTimeBeforeLunchToShipOrder,
		firstReminder:                      firstReminder,
		secondReminder:                     secondReminder,
	}
}

func (u *UsersReminder) Remind(ctx context.Context) {
	logrus.Info("usersReminder producer started")
	waitTimeToCreateTicker(ctx, u.startingMinutes)
	logrus.Infof("userReminder is ready to create ticker: %s", time.Now().UTC().String())
	t := time.NewTicker(u.tickInterval)
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("usersReminder producer stopped: %s", ctx.Err().Error())
			t.Stop()
			return
		case <-t.C:
			truncatedNowWithTimezone := time.Now().UTC().Add(u.timezone).Truncate(time.Minute)
			firstLunchTime := time.Duration(truncatedNowWithTimezone.Add(u.periodOfTimeBeforeLunchToShipOrder).Add(u.firstReminder).Hour())*time.Hour +
				time.Duration(truncatedNowWithTimezone.Add(u.periodOfTimeBeforeLunchToShipOrder).Add(u.firstReminder).Minute())*time.Minute
			secondLunchTime := time.Duration(truncatedNowWithTimezone.Add(u.periodOfTimeBeforeLunchToShipOrder).Add(u.secondReminder).Hour())*time.Hour +
				time.Duration(truncatedNowWithTimezone.Add(u.periodOfTimeBeforeLunchToShipOrder).Add(u.secondReminder).Minute())*time.Minute
			logrus.Debugf("remind: first lunch time: %s, second lunch time: %s",
				firstLunchTime.String(), secondLunchTime.String())
			lunchTimes := make([]string, 0)
			lunchTimes = append(lunchTimes, firstLunchTime.String(), secondLunchTime.String())

			newCtx, cancel := context.WithTimeout(ctx, time.Minute)
			telegramUsersByLunchTime, err := u.telegram.GetUsersByLunchTimes(newCtx, lunchTimes)
			if err != nil {
				logrus.Errorf("remind: %s", err.Error())
				cancel()
				continue
			}
			cancel()

			for lunchTime, telegramUsers := range telegramUsersByLunchTime {
				for _, tgUser := range telegramUsers {
					switch lunchTime {
					case firstLunchTime:
						msg := tgbotapi.NewMessage(tgUser.ChatID, fmt.Sprintf(firstOrderReminderMessage, u.firstReminder.Minutes()))
						_, err = u.bot.Send(msg)
						if err != nil {
							logrus.Errorf("remind: %s", err.Error())
							continue
						}
					case secondLunchTime:
						msg := tgbotapi.NewMessage(tgUser.ChatID, fmt.Sprintf(secondOrderReminderMessage, u.secondReminder.Minutes()))
						_, err = u.bot.Send(msg)
						if err != nil {
							logrus.Errorf("remind: %s", err.Error())
							continue
						}
					}
				}
			}
		}
	}
}

func waitTimeToCreateTicker(ctx context.Context, startingMinutes []int) {
	t := time.NewTicker(time.Second)
	var currentMinute int
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
			currentMinute = time.Now().Minute()
			for _, suitableMinute := range startingMinutes {
				if currentMinute == suitableMinute {
					t.Stop()
					return
				}
			}
		}
	}
}
