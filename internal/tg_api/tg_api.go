package tg_api

import (
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/ratelimit"
)

const maxRetries = 5

type TgAPI interface {
	SendMessage(chatID int64, msg tgbotapi.Chattable) (tgbotapi.Message, error)
	SetWebhook(webhookURL string) error
	AnswerCallbackQuery(chatID int64, callbackID string) error
	DeleteMessage(chatID int64, config tgbotapi.DeleteMessageConfig) (err error)
}

type tgAPI struct {
	botAPI        *tgbotapi.BotAPI
	limiter       ratelimit.Limiter
	limiterByChat sync.Map
}

type chatLimiter struct {
	limiter      ratelimit.Limiter
	lastActivity time.Time
}

func NewTgAPI(token string, clearPeriod time.Duration, chatTtl time.Duration) (TgAPI, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	tg := &tgAPI{
		botAPI:        botAPI,
		limiter:       ratelimit.New(25),
		limiterByChat: sync.Map{},
	}

	go tg.clearOldLimiters(clearPeriod, chatTtl)

	return tg, nil
}

func (tg *tgAPI) getChatLimiter(chatID int64) ratelimit.Limiter {
	chatLim, ok := tg.limiterByChat.Load(chatID)
	if !ok {
		chatLim = &chatLimiter{
			limiter:      ratelimit.New(2),
			lastActivity: time.Now(),
		}
		tg.limiterByChat.Store(chatID, chatLim)
	}
	return chatLim.(*chatLimiter).limiter
}

func (tg *tgAPI) withRetries(chatID int64, f func() error) {
	chatLimiter := tg.getChatLimiter(chatID)

	for i := 0; i < maxRetries; i++ {
		chatLimiter.Take()
		tg.limiter.Take()

		err := f()
		if err == nil {
			return
		}

		log.Printf("Попытка отправки сообщения %d: не удалось отправить запрос: %v", i+1, err)
		time.Sleep(time.Duration(1<<i) * time.Second)
	}
	return
}

func (tg *tgAPI) SendMessage(chatID int64, msg tgbotapi.Chattable) (sentMsg tgbotapi.Message, err error) {
	tg.withRetries(chatID, func() error {
		sentMsg, err = tg.botAPI.Send(msg)
		return err
	})
	return
}

func (tg *tgAPI) SetWebhook(webhookURL string) error {
	_, err := tg.botAPI.SetWebhook(tgbotapi.NewWebhook(webhookURL))
	return err
}

func (tg *tgAPI) AnswerCallbackQuery(chatID int64, callbackID string) (err error) {
	tg.withRetries(chatID, func() error {
		_, err = tg.botAPI.AnswerCallbackQuery(tgbotapi.NewCallback(callbackID, ""))
		return err
	})
	return
}

func (tg *tgAPI) DeleteMessage(chatID int64, config tgbotapi.DeleteMessageConfig) (err error) {
	tg.withRetries(chatID, func() error {
		_, err = tg.botAPI.DeleteMessage(config)
		return err
	})
	return err
}

func (tg *tgAPI) clearOldLimiters(clearPeriod time.Duration, chatTtl time.Duration) {
	for {
		time.Sleep(clearPeriod)
		now := time.Now()

		tg.limiterByChat.Range(func(key, value interface{}) bool {
			chatID := key.(int64)
			limiterData := value.(*chatLimiter)

			if now.Sub(limiterData.lastActivity) > chatTtl {
				tg.limiterByChat.Delete(chatID)
				log.Printf("clearOldLimiters: удален неактивный чат: %d", chatID)
			}
			return true
		})
	}
}
