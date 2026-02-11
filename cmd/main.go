// cmd/bot/main.go
package main

import (
	"fmt"
	"log"
	"strings"

	"nextturn/internal/config"
	"nextturn/internal/core"
	"nextturn/internal/db"
	"nextturn/internal/games/fourxo"
	"nextturn/internal/games/tictactoe"
	"nextturn/internal/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	config.Validate()

	// Register game engines
	core.Register(&tictactoe.Engine{})
	core.Register(&fourxo.Engine{})

	// Connect to database
	database, err := db.Connect(config.DatabaseURL)
	fmt.Println("Connected to database", database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Init bot
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Init handlers
	inlineRouter := handlers.NewInlineRouter(database)
	tttHandlers := tictactoe.NewHandlers(database)
	fourxoHandlers := fourxo.NewHandlers(database)

	// Start polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("Bot started with long polling...")

	for update := range updates {
		go func(update tgbotapi.Update) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered from panic: %v", r)
				}
			}()

			switch {
			case update.Message != nil && update.Message.IsCommand():
				handleCommand(bot, update.Message)

			case update.InlineQuery != nil:
				inlineRouter.HandleInlineQuery(bot, update.InlineQuery)

			case update.ChosenInlineResult != nil:
				inlineRouter.HandleChosenInlineResult(bot, update.ChosenInlineResult)

			case update.CallbackQuery != nil:
				handleCallback(bot, update.CallbackQuery, tttHandlers, fourxoHandlers)
			}
		}(update)
	}
}

func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		handlers.HandleStart(bot, msg)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery, tttHandlers *tictactoe.Handlers, fourxoHandlers *fourxo.Handlers) {
	data := cb.Data

	if data == "noop" {
		resp := tgbotapi.NewCallback(cb.ID, "")
		bot.Request(resp)
		return
	}

	prefix := strings.SplitN(data, ":", 2)[0]
	switch prefix {
	case "ttt":
		tttHandlers.HandleCallback(bot, cb)
	case "4xo":
		fourxoHandlers.HandleCallback(bot, cb)
	default:
		resp := tgbotapi.NewCallback(cb.ID, "Unknown action")
		bot.Request(resp)
	}
}
