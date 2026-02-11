// internal/handlers/start.go
package handlers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleStart(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.InlineKeyboardButton{Text: "⭕ Tic Tac Toe", SwitchInlineQuery: strPtr("ttt")},
			tgbotapi.InlineKeyboardButton{Text: "🎯 Four XO", SwitchInlineQuery: strPtr("4xo")},
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.InlineKeyboardButton{Text: "🐘 Elephant XO", SwitchInlineQuery: strPtr("exo")},
		),
	)

	text := "🎮 Welcome to Next Turn!\n\n" +
		"Choose a game and challenge your friends:\n\n" +
		"• Tic Tac Toe - Classic 3×3\n" +
		"• Four XO - 5×5 board, 4 in a row wins!\n" +
		"• Elephant XO - 7×7 board, 5 in a row wins!"

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ReplyMarkup = kb

	if _, err := bot.Send(reply); err != nil {
		log.Printf("HandleStart error: %v", err)
	}
}

func strPtr(s string) *string {
	return &s
}
