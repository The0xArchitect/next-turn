// cmd/main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"nextturn/internal/config"
	"nextturn/internal/core"
	"nextturn/internal/db"
	"nextturn/internal/games/checkers"
	"nextturn/internal/games/connect4"
	"nextturn/internal/games/elephantxo"
	"nextturn/internal/games/fourxo"
	"nextturn/internal/games/poolcheckers"
	"nextturn/internal/games/tictactoe"
	"nextturn/internal/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot            *tgbotapi.BotAPI
	database       *db.DB
	inlineRouter   *handlers.InlineRouter
	tttHandlers    *tictactoe.Handlers
	fourxoHandlers *fourxo.Handlers
	exoHandlers    *elephantxo.Handlers
	c4Handlers     *connect4.Handlers
	ckHandlers     *checkers.Handlers
	pcHandlers     *poolcheckers.Handlers
)

func init() {
	config.Validate()

	// Register game engines
	core.Register(&tictactoe.Engine{})
	core.Register(&fourxo.Engine{})
	core.Register(&elephantxo.Engine{})
	core.Register(&connect4.Engine{})
	core.Register(&checkers.Engine{})
	core.Register(&poolcheckers.Engine{})

	// Connect to database
	var err error
	database, err = db.Connect(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database")

	// Init bot
	bot, err = tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Init handlers
	inlineRouter = handlers.NewInlineRouter(database)
	tttHandlers = tictactoe.NewHandlers(database)
	fourxoHandlers = fourxo.NewHandlers(database)
	exoHandlers = elephantxo.NewHandlers(database)
	c4Handlers = connect4.NewHandlers(database)
	ckHandlers = checkers.NewHandlers(database)
	pcHandlers = poolcheckers.NewHandlers(database)
}

func main() {
	// Get port from environment variable (Cloud Run sets this)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up webhook handler
	http.HandleFunc("/", webhookHandler)
	http.HandleFunc("/health", healthHandler)

	log.Printf("Starting webhook server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("Recovered from panic: %v", rec)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("Failed to decode update: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process update asynchronously
	go processUpdate(update)

	// Respond immediately to Telegram
	w.WriteHeader(http.StatusOK)
}

func processUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in processUpdate: %v", r)
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
		handleCallback(bot, update.CallbackQuery, tttHandlers, fourxoHandlers, exoHandlers, c4Handlers, ckHandlers, pcHandlers)
	}
}

func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		handlers.HandleStart(bot, msg)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery,
	tttHandlers *tictactoe.Handlers,
	fourxoHandlers *fourxo.Handlers,
	exoHandlers *elephantxo.Handlers,
	c4Handlers *connect4.Handlers,
	ckHandlers *checkers.Handlers,
	pcHandlers *poolcheckers.Handlers,
) {
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
	case "exo":
		exoHandlers.HandleCallback(bot, cb)
	case "c4":
		c4Handlers.HandleCallback(bot, cb)
	case "ck":
		ckHandlers.HandleCallback(bot, cb)
	case "pc":
		pcHandlers.HandleCallback(bot, cb)
	default:
		resp := tgbotapi.NewCallback(cb.ID, "Unknown action")
		bot.Request(resp)
	}
}
