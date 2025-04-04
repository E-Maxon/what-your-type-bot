package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/E-Maxon/what-your-type-bot/config"
	"github.com/E-Maxon/what-your-type-bot/internal/tg_api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	maxAnswerWeight     = 3
	startCmd            = "start"
	backCmd             = "back"
	answerVariants      = []string{"Да", "Нет", "Частично"}
	answerVariantPoints = map[string]int{
		"Нет":      1,
		"Частично": 2,
		"Да":       3,
	}
)

type Bot interface {
	Start() error
}

type bot struct {
	tgAPI tg_api.TgAPI
	chats sync.Map
	cfg   *config.Config
}

type messageInfo struct {
	messageID   int
	messageText string
}

type chat struct {
	id             int64
	questionIndex  int
	answers        []string
	messageInfos   []messageInfo
	lastMessageID  int
	buttonsRemoved bool
	lastActivity   time.Time
}

func NewBot(cfg *config.Config) Bot {
	return &bot{
		cfg:   cfg,
		chats: sync.Map{},
	}
}

func (b *bot) Start() error {
	var err error
	b.tgAPI, err = tg_api.NewTgAPI(b.cfg.TelegramInfo.Token, b.cfg.ClearPeriod, b.cfg.ChatTtl)
	if err != nil {
		return err
	}

	err = b.tgAPI.SetWebhook(b.cfg.TelegramInfo.WebhookUrl)
	if err != nil {
		return err
	}

	go b.clearOldChats()

	http.HandleFunc("/", b.handleUpdates)
	return http.ListenAndServe(":8080", nil)
}

func (b *bot) handleUpdates(w http.ResponseWriter, r *http.Request) {
	var update tgbotapi.Update
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		log.Println("Error decoding update:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	if update.Message != nil {
		if update.Message.Text == "/start" {
			err := b.startQuiz(update.Message.Chat.ID, update.Message.From.ID)
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

func (b *bot) removeOldButtonsAndSend(chatID int64, msg tgbotapi.Chattable) (tgbotapi.Message, error) {
	chat, ok := b.getChat(chatID)
	if ok && !chat.buttonsRemoved {
		removeButtonsMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, chat.lastMessageID, tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{}))
		_, err := b.tgAPI.SendMessage(chatID, removeButtonsMsg)
		if err != nil {
			return tgbotapi.Message{}, err
		}
	}
	sentMsg, err := b.tgAPI.SendMessage(chatID, msg)
	if err != nil {
		return tgbotapi.Message{}, err
	}
	if ok {
		chat.buttonsRemoved = false
	}
	return sentMsg, nil
}

func (b *bot) startQuiz(chatID int64, userID int) error {
	msg := tgbotapi.NewMessage(chatID, b.cfg.QuizData.Greeting)
	msg.ReplyMarkup = createStartButton()
	sentMsg, err := b.removeOldButtonsAndSend(chatID, msg)
	if err != nil {
		return fmt.Errorf("startQuiz: Got error. ChatID: %d; UserID: %d; Error: %v", chatID, userID, err)
	}

	b.chats.Delete(chatID)
	b.chats.Store(chatID, &chat{
		id:             chatID,
		questionIndex:  0,
		answers:        []string{},
		lastMessageID:  sentMsg.MessageID,
		buttonsRemoved: false,
		lastActivity:   time.Now(),
	})
	return nil
}

func (b *bot) createButtons(chat *chat) [][]tgbotapi.InlineKeyboardButton {
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, variant := range answerVariants {
		keyboard = addButton(keyboard, variant, formatCallbackData(chat.questionIndex, variant))
	}
	if chat.questionIndex != 0 {
		keyboard = addButton(keyboard, "Назад", formatCallbackData(chat.questionIndex, backCmd))
	}
	keyboard = addButton(keyboard, "Сбросить результаты", &startCmd)
	return keyboard
}

func (b *bot) sendQuestion(chatID int64) error {
	chat, err := b.getChatOrErr(chatID)
	if err != nil {
		return fmt.Errorf("sendQuestion: %v", err)
	}
	if chat.questionIndex >= len(b.cfg.QuizData.Questions) {
		return fmt.Errorf("sendQuestion: there is no question with index %d", chat.questionIndex)
	}

	curQuestion := b.cfg.QuizData.Questions[chat.questionIndex]
	question := formatQuestion(chat.questionIndex, curQuestion)
	msg := tgbotapi.NewMessage(chatID, question)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(b.createButtons(chat)...)
	msg.ParseMode = "HTML"

	sentMsg, err := b.removeOldButtonsAndSend(chatID, msg)
	if err != nil {
		return fmt.Errorf("sendQuestion: %v", err)
	}

	chat.lastMessageID = sentMsg.MessageID
	chat.messageInfos = append(chat.messageInfos, messageInfo{
		messageID:   sentMsg.MessageID,
		messageText: question,
	})
	return nil
}

func (b *bot) sendResults(chatID int64) error {
	scores := map[string]*int{}
	chat, err := b.getChatOrErr(chatID)
	if err != nil {
		return fmt.Errorf("sendResults: %v", err)
	}

	for id, info := range b.cfg.QuizData.Calculation {
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
			*score += answerVariantPoints[chat.answers[questionIndex]]
		}
	}

	res := []*typeScore{}

	for id, score := range scores {
		res = append(res, &typeScore{
			id:    id,
			score: *score,
			desc:  b.cfg.QuizData.Calculation[id].Description,
			total: maxAnswerWeight * len(b.cfg.QuizData.Calculation[id].QuestionIndexes),
		})
	}

	slices.SortFunc(res, func(a, b *typeScore) int {
		if a.score > b.score {
			return -1
		}
		if a.score < b.score {
			return 1
		}
		return 0
	})

	text, err := formatResultsHeader(res)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	sentMsg, err := b.removeOldButtonsAndSend(chatID, msg)
	if err != nil {
		return err
	}
	chat.lastMessageID = sentMsg.MessageID
	chat.buttonsRemoved = true

	for i, score := range res {
		msg := tgbotapi.NewMessage(chatID, formatResultType(score))
		msg.ParseMode = "HTML"
		if i == len(res)-1 {
			keyboard := [][]tgbotapi.InlineKeyboardButton{}
			keyboard = addButton(keyboard, "Пройти тест заново", &startCmd)
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
		}
		sentMsg, err := b.removeOldButtonsAndSend(chatID, msg)
		if err != nil {
			return err
		}
		chat.lastMessageID = sentMsg.MessageID
		if i != len(res)-1 {
			chat.buttonsRemoved = true
		}
	}

	return nil
}

func (b *bot) handleCallback(callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID
	data := callback.Data // Ответ пользователя

	err := b.tgAPI.AnswerCallbackQuery(chatID, callback.ID)
	if err != nil {
		return err
	}

	chat, ok := b.getChat(chatID)
	if !ok {
		return b.startQuiz(chatID, callback.From.ID)
	}

	chat.lastActivity = time.Now()
	if data == startCmd {
		chat.answers = []string{}
		chat.messageInfos = []messageInfo{}
		chat.questionIndex = 0
		err = b.sendQuestion(chatID)
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

	if questionIndex != chat.questionIndex-1 {
		return nil
	}

	if answer == backCmd {
		err := b.tgAPI.DeleteMessage(chatID, tgbotapi.NewDeleteMessage(
			chatID,
			chat.messageInfos[chat.questionIndex-1].messageID,
		))
		if err != nil {
			return err
		}
		lastMessageInfo := chat.messageInfos[chat.questionIndex-2]
		msg := tgbotapi.NewEditMessageText(
			chatID,
			lastMessageInfo.messageID,
			lastMessageInfo.messageText,
		)
		chat.questionIndex -= 2
		msg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: b.createButtons(chat),
		}
		msg.ParseMode = "HTML"
		sentMsg, err := b.tgAPI.SendMessage(chatID, msg)
		if err != nil {
			return err
		}
		chat.lastMessageID = sentMsg.MessageID
		chat.answers = chat.answers[:len(chat.answers)-1]
		chat.messageInfos = chat.messageInfos[:len(chat.messageInfos)-1]
		chat.questionIndex++
		return nil
	}

	chat.answers = append(chat.answers, answer)

	if len(chat.messageInfos) > 0 {
		lastMessageInfo := chat.messageInfos[len(chat.messageInfos)-1]
		msg := tgbotapi.NewEditMessageText(
			chatID,
			lastMessageInfo.messageID,
			fmt.Sprintf("%s\n\n<b>Ваш ответ: %s</b>", lastMessageInfo.messageText, chat.answers[len(chat.answers)-1]),
		)
		msg.ParseMode = "HTML"
		_, err := b.tgAPI.SendMessage(chatID, msg)
		if err != nil {
			return err
		}
		chat.buttonsRemoved = true
	}

	// Переходим к следующему вопросу
	if chat.questionIndex < len(b.cfg.QuizData.Questions) {
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
	}

	return nil
}

func (b *bot) getChat(chatID int64) (*chat, bool) {
	res, ok := b.chats.Load(chatID)
	if !ok {
		return nil, false
	}
	return res.(*chat), true
}

func (b *bot) getChatOrErr(chatID int64) (*chat, error) {
	res, ok := b.getChat(chatID)
	if !ok {
		return nil, fmt.Errorf("chat with id %d not found", chatID)
	}
	return res, nil
}

func (b *bot) clearOldChats() {
	for {
		time.Sleep(b.cfg.ClearPeriod)
		now := time.Now()

		b.chats.Range(func(key, value interface{}) bool {
			chatID := key.(int64)
			chat := value.(*chat)

			if now.Sub(chat.lastActivity) > b.cfg.ChatTtl {
				b.chats.Delete(chatID)
				log.Printf("clearOldChats: удален неактивный чат: %d", chatID)
			}
			return true
		})
	}
}
