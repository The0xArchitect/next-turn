// internal/games/connect4/ui.go
package connect4

import (
	"fmt"

	"nextturn/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const prefix = "c4"

func cellDisplay(cell byte, isWinCell bool) string {
	if isWinCell {
		if cell == 'X' {
			return "🔷"
		}
		return "♦️"
	}
	switch cell {
	case 'X':
		return "🔵"
	case 'O':
		return "🔴"
	default:
		return "⚪"
	}
}

func LobbyWaiting(gameID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎮 Join Game", fmt.Sprintf("%s:join:%s", prefix, gameID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", fmt.Sprintf("%s:quit:%s", prefix, gameID)),
		),
	)
}

func LobbyReady(gameID string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("▶️ Start Game", fmt.Sprintf("%s:start:%s", prefix, gameID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Leave", fmt.Sprintf("%s:quit:%s", prefix, gameID)),
		),
	)
}

func Board(gameID, boardState string, frozen bool, winPositions map[int]bool) tgbotapi.InlineKeyboardMarkup {
	var kbRows [][]tgbotapi.InlineKeyboardButton
	for row := 0; row < rows; row++ {
		var buttons []tgbotapi.InlineKeyboardButton
		for col := 0; col < cols; col++ {
			pos := row*cols + col
			cell := boardState[pos]
			isWin := winPositions[pos]
			display := cellDisplay(cell, isWin)

			cbData := "noop"
			// Tap anywhere in column → drop disc there; column playable if top cell empty
			if !frozen && boardState[col] == '-' {
				cbData = fmt.Sprintf("%s:move:%s:%d", prefix, gameID, col)
			}
			buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(display, cbData))
		}
		kbRows = append(kbRows, buttons)
	}
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: kbRows}
}

func Finished(game *models.Game) tgbotapi.InlineKeyboardMarkup {
	winPositions := make(map[int]bool)
	if wp := FindWinningPositions(game.Board); wp != nil {
		for _, p := range wp {
			winPositions[p] = true
		}
	}
	kb := Board(game.ID, game.Board, true, winPositions)
	kb.InlineKeyboard = append(kb.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔄 Rematch?", fmt.Sprintf("%s:rematch:%s", prefix, game.ID)),
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
		return Board(game.ID, game.Board, false, nil)
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
		return "🔴🔵 Connect 4\n\n⏰ This game has expired"
	}
	return fmt.Sprintf("🔴🔵 Connect 4\n\n%s wants to play!\nTap Join to accept the challenge.", game.CreatorName)
}

func messageReady(game *models.Game) string {
	if game.IsExpired() {
		return "🔴🔵 Connect 4\n\n⏰ This game has expired"
	}
	return fmt.Sprintf("🔴🔵 Connect 4\n\n🔵 %s vs 🔴 %s\n\nPlayers ready! Host can start the game.",
		game.XPlayerName(), game.OPlayerName())
}

func messagePlaying(game *models.Game) string {
	if game.IsExpired() {
		return "🔴🔵 Connect 4\n\n⏰ This game has expired"
	}
	turnSymbol := "🔵"
	if game.CurrentTurn == "O" {
		turnSymbol = "🔴"
	}
	return fmt.Sprintf("🔴🔵 Connect 4\n\n🔵 %s vs 🔴 %s\n\n%s %s's turn - drop a disc!",
		game.XPlayerName(), game.OPlayerName(), turnSymbol, game.CurrentPlayerName())
}

func messageFinished(game *models.Game) string {
	var resultText string
	switch game.GetResult() {
	case "creator":
		resultText = fmt.Sprintf("🏆 %s wins! Matched 4 discs!", game.CreatorName)
	case "opponent":
		resultText = fmt.Sprintf("🏆 %s wins! Matched 4 discs!", game.GetOpponentName())
	case "draw":
		resultText = "🤝 It's a draw! The board is full."
	case "forfeit_creator":
		resultText = fmt.Sprintf("🚪 %s left. %s wins by default!", game.CreatorName, game.GetOpponentName())
	case "forfeit_opponent":
		resultText = fmt.Sprintf("🚪 %s left. %s wins by default!", game.GetOpponentName(), game.CreatorName)
	case "cancelled":
		resultText = "❌ Game was cancelled"
	default:
		resultText = "Game over"
	}
	return fmt.Sprintf("🔴🔵 Connect 4\n\n🔵 %s vs 🔴 %s\n\n%s",
		game.XPlayerName(), game.OPlayerName(), resultText)
}
