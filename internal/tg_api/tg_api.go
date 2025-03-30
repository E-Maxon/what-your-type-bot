package tg_api

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const maxRetries = 5

type TgAPI interface {
	SendMessage(chatID int64, msg tgbotapi.MessageConfig) error
	SetWebhook(webhookURL string) error
	ListenForWebhook(pattern string) tgbotapi.UpdatesChannel
}

type tgAPI struct {
	botAPI *tgbotapi.BotAPI
}

func NewTgAPI(token string) (TgAPI, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &tgAPI{botAPI: botAPI}, nil
}

func (tg *tgAPI) SendMessage(chatID int64, msg tgbotapi.MessageConfig) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		_, err = tg.botAPI.Send(msg)
		if err == nil {
			return nil
		}

		log.Printf("Попытка отправки сообщения %d: не удалось отправить запрос: %v", i+1, err)
		time.Sleep(time.Duration(2^i) * time.Second)
	}
	return err
}

func (tg *tgAPI) SetWebhook(webhookURL string) error {
	_, err := tg.botAPI.SetWebhook(tgbotapi.NewWebhook(webhookURL))
	return err
}

func (tg *tgAPI) ListenForWebhook(pattern string) tgbotapi.UpdatesChannel {
	return tg.botAPI.ListenForWebhook(pattern)
}
