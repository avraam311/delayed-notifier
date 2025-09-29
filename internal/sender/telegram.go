package sender

import (
	"encoding/json"

	"github.com/avraam311/delayed-notifier/internal/models"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	bot *tg.BotAPI
}

func NewBot(botToken string) (*TelegramBot, error) {
	bot, err := tg.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}
	return &TelegramBot{
		bot: bot,
	}, nil
}

func (tb *TelegramBot) SendMessage(msg []byte) error {
	var not *models.Notification
	json.Unmarshal(msg, &not)

	chatID := int64(not.UserID)

	msgToSend := tg.NewMessage(chatID, "Привет! Это тестовое сообщение из Go.")

	_, err := tb.bot.Send(msgToSend)
	if err != nil {
		return err
	}
	return nil
}
