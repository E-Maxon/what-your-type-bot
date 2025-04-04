package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type typeScore struct {
	id    string
	desc  string
	score int
	total int
}

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

func formatResultsHeader(scores []*typeScore) (res string, err error) {
	if len(scores) == 0 {
		return "", fmt.Errorf("must be 1 or more pyschotypes")
	}
	res += "<b>Результаты:</b>\n\n"
	res += fmt.Sprintf("Больше всего баллов набрал тип личности <b>%s</b>\n\n", scores[0].id)
	res += "<b>Набранные баллы:</b>"
	return res, nil
}

func formatResultType(score *typeScore) string {
	return fmt.Sprintf("<b>%s - %d/%d</b>\n\n%s", score.id, score.score, score.total, score.desc)
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
