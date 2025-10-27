package sender

import (
	"encoding/json"
	"strconv"

	"github.com/avraam311/delayed-notifier/internal/models/domain"

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
	var not *domain.Notification
	err := json.Unmarshal(msg, &not)
	if err != nil {
		return err
	}
	chatIDINT, err := strconv.Atoi(not.TgID)
	if err != nil {
		return err
	}
	chatID := int64(chatIDINT)
	text := not.Message

	msgToSend := tg.NewMessage(chatID, text)

	_, err = tb.bot.Send(msgToSend)
	if err != nil {
		return err
	}
	return nil
}
