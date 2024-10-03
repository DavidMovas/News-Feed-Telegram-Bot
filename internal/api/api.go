package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"telbot/internal/bot"
	"telbot/internal/botkit"
	"telbot/internal/config"
	"telbot/internal/fetcher"
	"telbot/internal/notifier"
	"telbot/internal/storage"
	"telbot/internal/summary"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type API struct {
	ctx context.Context
}

func New(ctx context.Context) *API {
	return &API{
		ctx: ctx,
	}
}

func (a *API) Run() error {
	botApi, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)

	if err != nil {
		return errors.New(fmt.Sprintf("[ERROR] failed to create bot api: %v", err))
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)

	if err != nil {
		return errors.New(fmt.Sprintf("[ERROR] failed to connect to database: %v", err))
	}
	defer db.Close()

	if err := db.PingContext(a.ctx); err != nil {
		return errors.New(fmt.Sprintf("[ERROR] failed to ping database: %v", err))
	}

	newsBot := botkit.New(botApi)
	newsBot.RegisterCmdView("start", bot.ViewCmdStart())

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

	ctx, cancel := signal.NotifyContext(a.ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func(ctx context.Context) {
		if err := fetcher.Start(ctx); err != nil {
			log.Printf("[ERROR] failed to start fetcher: %v", err)
			return
		}
		log.Println("[INFO] fetcher stopped")
	}(ctx)

	go func(ctx context.Context) {
		if err := notifier.Start(ctx); err != nil {
			log.Printf("[ERROR] failed to start notifier: %v", err)
		}
		log.Println("[INFO] notifier stopped")
	}(ctx)

	if err := newsBot.Run(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("[ERROR] failed to run bot: %v", err)
			return err
		}
	}

	return nil
}
