package tg_api

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const maxRetries = 5

type TgAPI interface {
	SendMessage(chatID int64, msg tgbotapi.Chattable) (tgbotapi.Message, error)
	SetWebhook(webhookURL string) error
	AnswerCallbackQuery(callbackID string) error
	DeleteMessage(config tgbotapi.DeleteMessageConfig) (err error)
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

func withRetries(f func() error) {
	for i := 0; i < maxRetries; i++ {
		err := f()
		if err == nil {
			return
		}

		log.Printf("Попытка отправки сообщения %d: не удалось отправить запрос: %v", i+1, err)
		time.Sleep(time.Duration(2^i) * time.Second)
	}
	return
}

func (tg *tgAPI) SendMessage(chatID int64, msg tgbotapi.Chattable) (sentMsg tgbotapi.Message, err error) {
	withRetries(func() error {
		sentMsg, err = tg.botAPI.Send(msg)
		return err
	})
	return
}

func (tg *tgAPI) SetWebhook(webhookURL string) error {
	_, err := tg.botAPI.SetWebhook(tgbotapi.NewWebhook(webhookURL))
	return err
}

func (tg *tgAPI) AnswerCallbackQuery(callbackID string) (err error) {
	withRetries(func() error {
		_, err = tg.botAPI.AnswerCallbackQuery(tgbotapi.NewCallback(callbackID, ""))
		return err
	})
	return
}

func (tg *tgAPI) DeleteMessage(config tgbotapi.DeleteMessageConfig) (err error) {
	withRetries(func() error {
		_, err = tg.botAPI.DeleteMessage(config)
		return err
	})
	return err
}
