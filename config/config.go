package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type QuizData struct {
	Greeting    string                `json:"greeting"`
	Questions   []string              `json:"questions"`
	Calculation map[string]*PsyhoType `json:"calculation"`
}

type PsyhoType struct {
	Description     string `json:"description"`
	QuestionIndexes []int  `json:"questions"`
}

type TelegramInfo struct {
	Token      string
	WebhookUrl string
}

type Config struct {
	QuizData     *QuizData
	TelegramInfo *TelegramInfo
}

func ParseConfig() (*Config, error) {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		return nil, fmt.Errorf("config path is empty")
	}
	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var quizData QuizData
	err = json.Unmarshal(file, &quizData)
	if err != nil {
		return nil, err
	}

	for _, psychoType := range quizData.Calculation {
		for i := range psychoType.QuestionIndexes {
			psychoType.QuestionIndexes[i]--
		}
	}

	cfg := &Config{
		QuizData: &quizData,
		TelegramInfo: &TelegramInfo{
			Token:      os.Getenv("BOT_TOKEN"),
			WebhookUrl: os.Getenv("WEBHOOK_URL"),
		},
	}

	return cfg, nil
}
