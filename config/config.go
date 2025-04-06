package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
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
	ClearPeriod  time.Duration
	ChatTtl      time.Duration
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

	var cfg struct {
		QuizData    *QuizData `json:"quiz_data"`
		ClearPeriod string    `json:"clear_period"`
		ChatTtl     string    `json:"chat_ttl"`
	}

	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}

	for _, psychoType := range cfg.QuizData.Calculation {
		for i := range psychoType.QuestionIndexes {
			psychoType.QuestionIndexes[i]--
		}
	}

	clearPeriod, err := time.ParseDuration(cfg.ClearPeriod)
	if err != nil {
		return nil, err
	}

	chatTtl, err := time.ParseDuration(cfg.ChatTtl)
	if err != nil {
		return nil, err
	}

	res := &Config{
		QuizData: cfg.QuizData,
		TelegramInfo: &TelegramInfo{
			Token:      os.Getenv("BOT_TOKEN"),
			WebhookUrl: os.Getenv("WEBHOOK_URL"),
		},
		ClearPeriod: clearPeriod,
		ChatTtl:     chatTtl,
	}

	return res, nil
}
