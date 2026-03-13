// internal/handlers/inline_router.go
package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"nextturn/internal/core"
	"nextturn/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var queryAliases = map[string]string{
	"ttt":          "tictactoe",
	"tic":          "tictactoe",
	"tictactoe":    "tictactoe",
	"4xo":          "fourxo",
	"fourxo":       "fourxo",
	"four":         "fourxo",
	"exo":          "elephantxo",
	"e5":           "elephantxo",
	"elephant":     "elephantxo",
	"elephantxo":   "elephantxo",
	"c4":           "connect4",
	"connect4":     "connect4",
	"ck":           "checkers",
	"checkers":     "checkers",
	"pc":           "poolcheckers",
	"pool":         "poolcheckers",
	"poolcheckers": "poolcheckers",
}

const thumbBaseURL = "https://alice-bikeshop.web.app/next-turn/"

type InlineRouter struct {
	DB *db.DB
}

func NewInlineRouter(database *db.DB) *InlineRouter {
	return &InlineRouter{DB: database}
}

func (r *InlineRouter) HandleInlineQuery(bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) {
	q := strings.TrimSpace(strings.ToLower(query.Query))
	user := query.From

	gameType, ok := queryAliases[q]
	if q == "" || !ok {
		r.showGameMenu(bot, query, user)
	} else {
		r.showGamePreview(bot, query, user, gameType)
	}
}

func (r *InlineRouter) showGameMenu(bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery, user *tgbotapi.User) {
	var results []interface{}

	for _, engine := range core.All() {
		resultID := fmt.Sprintf("new:%s:%d", engine.GameType(), user.ID)
		text := buildPreviewMessage(engine, user)
		kb := buildPreviewKeyboard()

		article := tgbotapi.NewInlineQueryResultArticleHTML(resultID, engine.DisplayName(), text)
		article.Description = engine.Description()
		article.ReplyMarkup = &kb
		article.ThumbURL = thumbBaseURL + engine.ThumbnailURL()

		results = append(results, article)
	}

	answer := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       results,
		CacheTime:     0,
		IsPersonal:    true,
	}
	if _, err := bot.Request(answer); err != nil {
		log.Printf("showGameMenu error: %v", err)
	}
}

func (r *InlineRouter) showGamePreview(bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery, user *tgbotapi.User, gameType string) {
	engine := core.Get(gameType)
	if engine == nil {
		r.showGameMenu(bot, query, user)
		return
	}

	resultID := fmt.Sprintf("new:%s:%d", engine.GameType(), user.ID)
	text := buildPreviewMessage(engine, user)
	kb := buildPreviewKeyboard()

	article := tgbotapi.NewInlineQueryResultArticleHTML(resultID, engine.DisplayName(), text)
	article.Description = "Tap to send"
	article.ReplyMarkup = &kb
	article.ThumbURL = thumbBaseURL + engine.ThumbnailURL()

	answer := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       []interface{}{article},
		CacheTime:     0,
		IsPersonal:    true,
	}
	if _, err := bot.Request(answer); err != nil {
		log.Printf("showGamePreview error: %v", err)
	}
}

func buildPreviewMessage(engine core.GameEngine, user *tgbotapi.User) string {
	return fmt.Sprintf("🎮 %s\n\n%s wants to play!\n\nWaiting for opponent...",
		engine.DisplayName(), user.FirstName)
}

func buildPreviewKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏳ Loading...", "noop"),
		),
	)
}

func (r *InlineRouter) HandleChosenInlineResult(bot *tgbotapi.BotAPI, chosen *tgbotapi.ChosenInlineResult) {
	resultID := chosen.ResultID
	inlineMessageID := chosen.InlineMessageID
	user := chosen.From

	if inlineMessageID == "" {
		return
	}
	if !strings.HasPrefix(resultID, "new:") {
		return
	}

	parts := strings.Split(resultID, ":")
	if len(parts) < 3 {
		return
	}

	gameType := parts[1]
	engine := core.Get(gameType)
	if engine == nil {
		return
	}

	ctx := context.Background()
	game, err := r.DB.CreateGame(ctx, user.ID, user.FirstName, gameType, engine.InitialBoard())
	if err != nil {
		log.Printf("CreateGame error: %v", err)
		return
	}

	_, err = r.DB.SetInlineMessageID(ctx, game.ID, inlineMessageID)
	if err != nil {
		log.Printf("SetInlineMessageID error: %v", err)
		return
	}

	text := engine.BuildMessage(game)
	kb := engine.BuildLobbyWaiting(game.ID)

	edit := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			InlineMessageID: inlineMessageID,
			ReplyMarkup:     &kb,
		},
		Text: text,
	}
	if _, err := bot.Request(edit); err != nil {
		log.Printf("editInlineMessage error: %v", err)
	}
}
