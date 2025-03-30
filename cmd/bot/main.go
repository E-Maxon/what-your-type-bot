package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/E-Maxon/what-your-type-bot/config"
	"github.com/E-Maxon/what-your-type-bot/internal/bot"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Ошибка при парсинге конфига: %v", err)
	}
	b := bot.NewBot(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := b.Start(ctx); err != nil {
			errCh <- fmt.Errorf("Ошибка при запуске бота: %v", err)
		}
	}()

	select {
	case <-stop:
		log.Println("Получен сигнал завершения. Ожидаем завершения работы бота...")
	case err := <-errCh:
		log.Printf("Произошла ошибка: %v", err)
	}

	time.Sleep(2 * time.Second)
	log.Println("Бот завершил свою работу. Выход.")
}
