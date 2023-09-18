package config

import (
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/sirupsen/logrus"
)

type Config struct {
	LogLevel                           int           `env:"LOG_LEVEL"`
	Timezone                           time.Duration `env:"TIMEZONE"`
	StartingMinutes                    []int         `env:"STARTING_MINUTES"`
	TickInterval                       time.Duration `env:"TICK_INTERVAL"`
	PeriodOfTimeBeforeLunchToShipOrder time.Duration `env:"PERIOD_OF_TIME_BEFORE_LUNCH_TO_SHIP_ORDER"`
	AdminChatID                        int64         `env:"ADMIN_CHAT_ID"`
	StartedLunchTime                   time.Duration `env:"STARTED_LUNCH_TIME"`
	FinishedLunchTime                  time.Duration `env:"FINISHED_LUNCH_TIME"`
	Postgres
	TelegramBot
	Menu
	UsersReminder
	StatisticsSender
}

type Postgres struct {
	DB       string `env:"POSTGRES_DB"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	Port     string `env:"POSTGRES_PORT"`
	Endpoint string `env:"POSTGRES_ENDPOINT"`
}

type TelegramBot struct {
	Token   string `env:"BOT_TOKEN"`
	Timeout int    `env:"BOT_TIMEOUT"`
}

type Menu struct {
	Categories []string           `env:"CATEGORIES"`
	Soups      map[string]float32 `env:"SOUPS"`
	Salads     map[string]float32 `env:"SALADS"`
	MainCourse map[string]float32 `env:"MAIN_COURSE"`
	Desserts   map[string]float32 `env:"DESSERTS"`
	Drinks     map[string]float32 `env:"DRINKS"`
}

type UsersReminder struct {
	FirstReminder  time.Duration `env:"FIRST_USERS_REMINDER"`
	SecondReminder time.Duration `env:"SECOND_USERS_REMINDER"`
}

type StatisticsSender struct {
	ReportHour      int     `env:"REPORT_HOUR"`
	ReportReceivers []int64 `env:"REPORT_RECEIVERS"`
}

func NewConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.Fatalf("%+v\n", err)
	}
	return &cfg
}
