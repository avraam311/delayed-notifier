package config

import (
	"os"
	"strconv"

	"github.com/wb-go/wbf/config"
)

type DB struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type Redis struct {
	Address  string
	Password string
	Number   int
}

type SMTP struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type TG struct {
	Token  string
	ChatID int
}

type Env struct {
	DB    *DB
	Redis *Redis
	SMTP  *SMTP
	TG    *TG
}

type Config struct {
	Env *Env
	Cfg *config.Config
}

func MustLoad() (*Config, error) {
	cfg := config.New()
	err := cfg.Load("config/local.yaml")
	if err != nil {
		return nil, err
	}

	db_port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, err
	}
	db := DB{
		Host:     os.Getenv("DB_HOST"),
		Port:     db_port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
	}
	redis_number, err := strconv.Atoi(os.Getenv("REDIS_NUMBER"))
	if err != nil {
		return nil, err
	}
	redis := Redis{
		Address:  os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		Number:   redis_number,
	}
	smtp_port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, err
	}
	smtp := SMTP{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     smtp_port,
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMRT_From"),
	}
	tg_chat_id, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	if err != nil {
		return nil, err
	}
	tg := TG{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		ChatID: tg_chat_id,
	}
	env := Env{
		DB:    &db,
		Redis: &redis,
		SMTP:  &smtp,
		TG:    &tg,
	}
	return &Config{
		Env: &env,
		Cfg: cfg,
	}, nil
}
