package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Greeting     string                `json:"greeting"`
	Questions    []string              `json:"questions"`
	Calculation  map[string]*PsyhoType `json:"calculation"`
	TelegramInfo *TelegramInfo         `json:"telegram"`
}

type TelegramInfo struct {
	Token      string `json:"token"`
	WebhookUrl string `json:"webhook_url"`
}

type PsyhoType struct {
	Description     string `json:"description"`
	QuestionIndexes []int  `json:"questions"`
}

func ParseConfig() (*Config, error) {
	file, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}

	for _, psychoType := range cfg.Calculation {
		for i := range psychoType.QuestionIndexes {
			psychoType.QuestionIndexes[i]--
		}
	}

	return &cfg, nil
}
