package main

import (
	"log"

	"github.com/E-Maxon/what-your-type-bot/config"
	"github.com/E-Maxon/what-your-type-bot/internal/bot"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Ошибка при парсинге конфига: %v", err)
	}
	b := bot.NewBot(cfg)

	if err := b.Start(); err != nil {
		log.Fatal(err)
	}
}
