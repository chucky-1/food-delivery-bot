package producer

import (
	"context"
	"fmt"
	"time"

	"github.com/chucky-1/food-delivery-bot/internal/model"
	"github.com/chucky-1/food-delivery-bot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const (
	dayPeriod   = "day"
	monthPeriod = "month"
)

type StatisticsSender struct {
	bot             *tgbotapi.BotAPI
	statistics      service.Statistics
	timezone        time.Duration
	reportHour      int
	reportReceivers []int64
}

func NewStatisticsSender(bot *tgbotapi.BotAPI, statistics service.Statistics, timezone time.Duration, reportHour int,
	reportReceivers []int64) *StatisticsSender {
	return &StatisticsSender{
		bot:             bot,
		statistics:      statistics,
		timezone:        timezone,
		reportHour:      reportHour,
		reportReceivers: reportReceivers,
	}
}

func (s *StatisticsSender) StatisticsSend(ctx context.Context) {
	logrus.Info("statisticSender producer started")
	waitTimeToCreateTickerForStatisticsSender(ctx)
	logrus.Info("statisticSender producer is ready to create ticker: %s", time.Now().UTC())
	t := time.NewTicker(time.Hour)
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("statisticSender producer stopped: %s", ctx.Err().Error())
			t.Stop()
			return
		case <-t.C:
			if time.Now().UTC().Add(s.timezone).Hour() == s.reportHour {
				err := s.sendReportsForYesterday(ctx)
				if err != nil {
					logrus.Errorf("statisticsSend: %s", err.Error())
				}
			}
			if time.Now().UTC().Add(s.timezone).Hour() == s.reportHour && time.Now().UTC().Add(s.timezone).Day() == 1 {
				err := s.sendReportsForMonth(ctx)
				if err != nil {
					logrus.Errorf("statisticsSend: %s", err.Error())
				}
			}
		}
	}
}

func (s *StatisticsSender) sendReportsForYesterday(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, time.Minute)
	stats, err := s.statistics.GetByDate(newCtx, time.Now().UTC().Add(s.timezone).Add(-24*time.Hour))
	if err != nil {
		cancel()
		return fmt.Errorf("sendReportsForYesterday: %w", err)
	}
	cancel()

	msg := s.createReportMessage(stats, dayPeriod)
	for _, chatID := range s.reportReceivers {
		tgMsg := tgbotapi.NewMessage(chatID, msg)
		_, err = s.bot.Send(tgMsg)
		if err != nil {
			return fmt.Errorf("send: %w", err)
		}
	}
	return nil
}

func (s *StatisticsSender) sendReportsForMonth(ctx context.Context) error {
	from, to := s.firstAndLastDaysOfPreviousMonth()
	logrus.Debugf("sendReportsForMonth: from: %s, to: %s", from.String(), to.String())

	newCtx, cancel := context.WithTimeout(ctx, time.Minute)
	stats, err := s.statistics.GetByDatesInterval(newCtx, from, to)
	if err != nil {
		cancel()
		return fmt.Errorf("sendReportsForMonth: %w", err)
	}
	cancel()

	msg := s.createReportMessage(stats, monthPeriod)
	for _, chatID := range s.reportReceivers {
		tgMsg := tgbotapi.NewMessage(chatID, msg)
		_, err = s.bot.Send(tgMsg)
		if err != nil {
			return fmt.Errorf("send: %w", err)
		}
	}
	return nil
}

func (s *StatisticsSender) createReportMessage(stats []*model.Statistics, period string) string {
	yesterday := time.Now().UTC().Add(s.timezone).Add(-24 * time.Hour).Truncate(24 * time.Hour)
	var (
		msg         string
		totalAmount float32
	)
	switch period {
	case dayPeriod:
		msg = fmt.Sprintf("Отчёт за %d %s\n", yesterday.Day(), translateMonthWithDeclination(yesterday.Month()))
	case monthPeriod:
		msg = fmt.Sprintf("Отчёт за %s\n", translateMonthWithoutDeclination(yesterday.Month()))
	}

	for _, st := range stats {
		msg = fmt.Sprintf("%s%s - %.2f\n", msg, st.OrganizationName, st.OrdersAmount)
		totalAmount += st.OrdersAmount
	}
	msg = fmt.Sprintf("%s\nИтого: %.2f", msg, totalAmount)
	return msg
}

func (s *StatisticsSender) firstAndLastDaysOfPreviousMonth() (time.Time, time.Time) {
	currentTime := time.Now().UTC().Add(s.timezone)
	firstDayOfCurrentMonth := time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDayOfPreviousMonth := firstDayOfCurrentMonth.Add(-time.Hour * 24)
	firstDayOfPreviousMonth := time.Date(lastDayOfPreviousMonth.Year(), lastDayOfPreviousMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	return firstDayOfPreviousMonth, lastDayOfPreviousMonth
}

func waitTimeToCreateTickerForStatisticsSender(ctx context.Context) {
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
			if time.Now().Minute() == 0 {
				t.Stop()
				return
			}
		}
	}
}

func translateMonthWithDeclination(month time.Month) string {
	switch month {
	case time.January:
		return "Сентября"
	case time.February:
		return "Февраля"
	case time.March:
		return "Марта"
	case time.April:
		return "Апреля"
	case time.May:
		return "Мая"
	case time.June:
		return "Июня"
	case time.July:
		return "Июля"
	case time.August:
		return "Августа"
	case time.September:
		return "Сентября"
	case time.October:
		return "Октября"
	case time.November:
		return "Ноября"
	case time.December:
		return "Декабря"
	}
	return ""
}

func translateMonthWithoutDeclination(month time.Month) string {
	switch month {
	case time.January:
		return "Сентябрь"
	case time.February:
		return "Февраль"
	case time.March:
		return "Март"
	case time.April:
		return "Апрель"
	case time.May:
		return "Май"
	case time.June:
		return "Июнь"
	case time.July:
		return "Июль"
	case time.August:
		return "Август"
	case time.September:
		return "Сентябрь"
	case time.October:
		return "Октябрь"
	case time.November:
		return "Ноябрь"
	case time.December:
		return "Декабрь"
	}
	return ""
}
