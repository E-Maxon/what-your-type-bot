package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func createStartButton() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Начать тест", startCmd),
		),
	)
}

func formatQuestion(index int, question string) string {
	return fmt.Sprintf("<b>Вопрос №%d:</b> %s", index+1, question)
}

func formatCallbackData(index int, variant string) *string {
	res := fmt.Sprintf("%d:%s", index, variant)
	return &res
}

func addButton(keyboard [][]tgbotapi.InlineKeyboardButton, text string, data *string) [][]tgbotapi.InlineKeyboardButton {
	return append(keyboard, []tgbotapi.InlineKeyboardButton{
		{
			Text:         text,
			CallbackData: data,
		},
	})
}

func parseCallbackData(data string) (int, string, error) {
	parts := strings.Split(data, ":")
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("parseCallbackData: must be 2 parts in callback data")
	}

	questionIndex, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("Atoi error: %v", err)
	}

	answer := parts[1]
	return questionIndex, answer, nil
}
