// internal/games/elephantxo/ui.go
package elephantxo

import (
	"fmt"

	"nextturn/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const prefix = "exo"

func cellDisplay(cell byte) string {
	switch cell {
	case 'X':
		return "❌"
	case 'O':
		return "⭕"
	default:
		return "\u3000" // ideographic space
	}
}

func LobbyWaiting(gameID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎮 Join", fmt.Sprintf("%s:join:%s", prefix, gameID)),
			tgbotapi.NewInlineKeyboardButtonData("🚪 Bail", fmt.Sprintf("%s:quit:%s", prefix, gameID)),
		),
	)
}

func LobbyReady(gameID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🚀 Let's Go!", fmt.Sprintf("%s:start:%s", prefix, gameID)),
			tgbotapi.NewInlineKeyboardButtonData("🚪 Quit", fmt.Sprintf("%s:quit:%s", prefix, gameID)),
		),
	)
}

func Board(gameID, boardState string, frozen bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for row := 0; row < size; row++ {
		var buttons []tgbotapi.InlineKeyboardButton
		for col := 0; col < size; col++ {
			pos := row*size + col
			cell := boardState[pos]
			display := cellDisplay(cell)
			cbData := "noop"
			if !frozen && cell == '-' {
				cbData = fmt.Sprintf("%s:move:%s:%d", prefix, gameID, pos)
			}
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(display, cbData))
		}
		rows = append(rows, buttons)
	}
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func Finished(game *models.Game) tgbotapi.InlineKeyboardMarkup {
	kb := Board(game.ID, game.Board, true)
	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔄 Play Again", fmt.Sprintf("%s:rematch:%s", prefix, game.ID)),
	))
	return kb
}

func Keyboard(game *models.Game) tgbotapi.InlineKeyboardMarkup {
	if game.IsExpired() && game.Status != "finished" {
		return tgbotapi.NewInlineKeyboardMarkup()
	}
	switch game.Status {
	case "waiting":
		return LobbyWaiting(game.ID)
	case "ready":
		return LobbyReady(game.ID)
	case "playing":
		return Board(game.ID, game.Board, false)
	case "finished":
		return Finished(game)
	default:
		return tgbotapi.NewInlineKeyboardMarkup()
	}
}

func Message(game *models.Game) string {
	switch game.Status {
	case "waiting":
		return messageWaiting(game)
	case "ready":
		return messageReady(game)
	case "playing":
		return messagePlaying(game)
	case "finished":
		return messageFinished(game)
	default:
		return "Unknown game state"
	}
}

func messageWaiting(game *models.Game) string {
	if game.IsExpired() {
		return "🐘 Elephant XO\n\n⏰ Game expired (48h limit)"
	}
	return fmt.Sprintf("🐘 Elephant XO (7×7, 5 to win)\n\n%s is looking for a challenger...\nWho dares? 👀", game.CreatorName)
}

func messageReady(game *models.Game) string {
	if game.IsExpired() {
		return "🐘 Elephant XO\n\n⏰ Game expired (48h limit)"
	}
	return fmt.Sprintf("🐘 Elephant XO (7×7, 5 to win)\n\n❌ %s vs ⭕ %s\n\nBoth players locked in! Let's gooo 🔥",
		game.XPlayerName(), game.OPlayerName())
}

func messagePlaying(game *models.Game) string {
	if game.IsExpired() {
		return "🐘 Elephant XO\n\n⏰ Game expired (48h limit)"
	}
	turnSymbol := "❌"
	if game.CurrentTurn == "O" {
		turnSymbol = "⭕"
	}
	return fmt.Sprintf("🐘 Elephant XO (7×7, 5 to win)\n\n❌ %s vs ⭕ %s\n\n%s %s's move 🐘",
		game.XPlayerName(), game.OPlayerName(), turnSymbol, game.CurrentPlayerName())
}

func messageFinished(game *models.Game) string {
	var resultText string
	switch game.GetResult() {
	case "creator":
		resultText = fmt.Sprintf("👑 %s wins! GG", game.CreatorName)
	case "opponent":
		resultText = fmt.Sprintf("👑 %s wins! GG", game.GetOpponentName())
	case "draw":
		resultText = "🤝 Draw! Evenly matched"
	case "forfeit_creator":
		resultText = fmt.Sprintf("🏃 %s ran away. %s takes it!", game.CreatorName, game.GetOpponentName())
	case "forfeit_opponent":
		resultText = fmt.Sprintf("🏃 %s ran away. %s takes it!", game.GetOpponentName(), game.CreatorName)
	case "cancelled":
		resultText = "❌ Game cancelled"
	default:
		resultText = "Game over"
	}
	return fmt.Sprintf("🐘 Elephant XO (7×7, 5 to win)\n\n❌ %s vs ⭕ %s\n\n%s",
		game.XPlayerName(), game.OPlayerName(), resultText)
}
