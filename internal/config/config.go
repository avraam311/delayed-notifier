package config

import (
	"os"

	"github.com/wb-go/wbf/config"
)

type Config struct {
	Env *Env
	Cfg *config.Config
}

type Env struct {
	BotToken string
}

func MustLoad() (*Config, error) {
	cfg := config.New()
	err := cfg.Load("config/local.yaml")
	if err != nil {
		return nil, err
	}
	env := &Env{}
	env.BotToken = os.Getenv("botToken")
	return &Config{
		Env: env,
		Cfg: cfg,
	}, nil
}
