package bot

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/E-Maxon/what-your-type-bot/config"
	"github.com/E-Maxon/what-your-type-bot/internal/tg_api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	startCmd       = "start"
	backCmd        = "back"
	answerVariants = map[string]int{
		"Нет":      1,
		"Частично": 2,
		"Да":       3,
	}
)

type Bot interface {
	Start(ctx context.Context) error
}

type bot struct {
	tgAPI tg_api.TgAPI
	chats map[int64]*chat
	cfg   *config.Config
}

type chat struct {
	id            int64
	questionIndex int
	answers       []int
}

func NewBot(cfg *config.Config) Bot {
	return &bot{
		cfg:   cfg,
		chats: make(map[int64]*chat),
	}
}

func (b *bot) Start(ctx context.Context) error {
	var err error
	b.tgAPI, err = tg_api.NewTgAPI(b.cfg.TelegramInfo.Token)
	if err != nil {
		return err
	}

	err = b.tgAPI.SetWebhook(b.cfg.TelegramInfo.WebhookUrl)
	if err != nil {
		return err
	}

	updates := b.tgAPI.ListenForWebhook("/")
	for {
		select {
		case <-ctx.Done():
			log.Println("Got stop signal. Stopping the bot")
			return nil
		case update := <-updates:
			if update.Message != nil {
				if update.Message.Text == "/start" {
					err := b.startQuiz(update.Message)
					if err != nil {
						log.Println(err)
					}
				}
			}

			if update.CallbackQuery != nil {
				err := b.handleCallback(update.CallbackQuery)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func (b *bot) startQuiz(update *tgbotapi.Message) error {
	chatID := update.Chat.ID
	msg := tgbotapi.NewMessage(chatID, b.cfg.Greeting)
	msg.ReplyMarkup = createStartButton()
	err := b.tgAPI.SendMessage(chatID, msg)
	if err != nil {
		return fmt.Errorf("startQuiz: Got error. ChatID: %d; UserID: %d; Error: %v", chatID, update.From.ID, err)
	}
	delete(b.chats, chatID)
	return nil
}

func (b *bot) sendQuestion(chatID int64) error {
	chat, err := b.getChat(chatID)
	if err != nil {
		return fmt.Errorf("sendQuestion: %v", err)
	}
	if chat.questionIndex >= len(b.cfg.Questions) {
		return fmt.Errorf("sendQuestion: there is no question with index %d", chat.questionIndex)
	}

	curQuestion := b.cfg.Questions[chat.questionIndex]

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for variant := range answerVariants {
		keyboard = addButton(keyboard, variant, formatCallbackData(chat.questionIndex, variant))
	}
	if chat.questionIndex != 0 {
		keyboard = addButton(keyboard, "Назад", formatCallbackData(chat.questionIndex, backCmd))
	}
	keyboard = addButton(keyboard, "Сбросить результаты и начать заново", &startCmd)

	msg := tgbotapi.NewMessage(chatID, formatQuestion(chat.questionIndex, curQuestion))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	return b.tgAPI.SendMessage(chatID, msg)
}

func (b *bot) sendResults(chatID int64) error {
	scores := map[string]*int{}
	chat, err := b.getChat(chatID)
	if err != nil {
		return fmt.Errorf("sendResults: %v", err)
	}

	for id, info := range b.cfg.Calculation {
		for _, questionIndex := range info.QuestionIndexes {
			if questionIndex >= len(chat.answers) {
				return fmt.Errorf("sendResults: questionIndex >= len(answers)")
			}
			score, ok := scores[id]
			if !ok {
				t := 0
				scores[id] = &t
				score = scores[id]
			}
			*score += chat.answers[questionIndex]
		}
	}

	res := []struct {
		id    string
		score int
	}{}

	for id, score := range scores {
		res = append(res, struct {
			id    string
			score int
		}{
			id:    id,
			score: *score,
		})
	}

	slices.SortFunc(res, func(a, b struct {
		id    string
		score int
	}) int {
		if a.score < b.score {
			return -1
		}
		if a.score > b.score {
			return 1
		}
		return 0
	})

	resType := res[len(res)-1]
	text := fmt.Sprintf("Результат:\nВаш тип личности - %s\nОписание типа: %s", resType.id, b.cfg.Calculation[resType.id].Description)

	msg := tgbotapi.NewMessage(chatID, text)
	return b.tgAPI.SendMessage(chatID, msg)
}

func (b *bot) handleCallback(callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID
	data := callback.Data // Ответ пользователя

	if _, ok := b.chats[chatID]; !ok {
		b.chats[chatID] = &chat{
			id:            chatID,
			questionIndex: 0,
			answers:       []int{},
		}
	}

	chat := b.chats[chatID]

	if data == startCmd {
		delete(b.chats, chatID)
		err := b.sendQuestion(chatID)
		if err != nil {
			return fmt.Errorf("got error. chatID: %d; userID: %d; method: sendQuestion; error: %v\n", chatID, callback.From.ID, err)
		}
		chat.questionIndex++
		return nil
	}

	questionIndex, answer, err := parseCallbackData(data)
	if err != nil {
		return fmt.Errorf("CallbackData parsing error: %v", err)
	}

	if questionIndex != chat.questionIndex {
		return nil
	}

	if answer == backCmd {
		chat.questionIndex -= 2
		err := b.sendQuestion(chatID)
		if err != nil {
			chat.questionIndex += 2
			return fmt.Errorf("got error. chatID: %d; userID: %d; method: sendQuestion; error: %v\n", chatID, callback.From.ID, err)
		}
		chat.questionIndex++
		return nil
	}

	chat.answers = append(chat.answers, answerVariants[answer])

	// Переходим к следующему вопросу
	if chat.questionIndex < len(b.cfg.Questions) {
		err := b.sendQuestion(chatID)
		if err != nil {
			return fmt.Errorf("got error. chatID: %d; userID: %d; method: sendQuestion; error: %v\n", chatID, callback.From.ID, err)
		}
		chat.questionIndex++
	} else {
		// Завершаем тест и отправляем результаты
		err := b.sendResults(chatID)
		if err != nil {
			return fmt.Errorf("got error. chatID: %d; userID: %d; method: sendQuestion; error: %v\n", chatID, callback.From.ID, err)
		}
		delete(b.chats, chatID)
	}

	return nil
}

func (b *bot) getChat(chatID int64) (*chat, error) {
	chat, ok := b.chats[chatID]
	if !ok {
		return nil, fmt.Errorf("can't find chat with id %d", chatID)
	}
	return chat, nil
}
