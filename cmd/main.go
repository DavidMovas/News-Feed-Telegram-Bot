package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"telbot/internal/config"
	"telbot/internal/fetcher"
	"telbot/internal/notifier"
	"telbot/internal/storage"
	"telbot/internal/summary"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

func main() {
	botApi, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)

	if err != nil {
		log.Printf("[ERROR] failed to create bot api: %v", err)
		return
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)

	if err != nil {
		log.Printf("[ERROR] failed to connect to database: %v", err)
		return
	}
	defer db.Close()

	var (
		articleStorage = storage.NewArticleStorage(db)
		sourceStorage  = storage.NewSourceStorage(db)
		fetcher        = fetcher.New(
			articleStorage,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)
		notifier = notifier.New(
			articleStorage,
			summary.NewOpenAISummarizer(config.Get().OpenAIKey, config.Get().OpenAIPrompt),
			botApi,
			config.Get().NotificationInterval,
			2*config.Get().FetchInterval,
			config.Get().TelegramChannelID,
		)
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func(ctx context.Context) {
		if err := fetcher.Start(ctx); err != nil {
			log.Printf("[ERROR] failed to start fetcher: %v", err)
			return
		}
		log.Println("[INFO] fetcher stopped")
	}(ctx)

	//go func(ctx context.Context) {
	if err := notifier.Start(ctx); err != nil {
		log.Printf("[ERROR] failed to start notifier: %v", err)
		return
	}
	log.Println("[INFO] notifier stopped")
	//}(ctx)
}
