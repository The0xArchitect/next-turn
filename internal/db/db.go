// internal/db/db.go
package db

import (
	"context"
	"database/sql"
	"fmt"

	"nextturn/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	conn *sql.DB
}

func Connect(databaseURL string) (*DB, error) {
	conn, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	conn.SetMaxOpenConns(10)
	return &DB{conn: conn}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

const cols = `id, inline_message_id, game_type, creator_id, creator_name,
	opponent_id, opponent_name, status, board, x_player, current_turn, selected_pos,
	result, created_at, updated_at`

const selectAll = `SELECT ` + cols + ` FROM games`

const returning = ` RETURNING ` + cols

func scanGame(row interface{ Scan(dest ...any) error }) (*models.Game, error) {
	g := &models.Game{}
	err := row.Scan(
		&g.ID, &g.InlineMessageID, &g.GameType,
		&g.CreatorID, &g.CreatorName,
		&g.OpponentID, &g.OpponentName,
		&g.Status, &g.Board, &g.XPlayer, &g.CurrentTurn, &g.SelectedPos,
		&g.Result, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func scanGameOrNil(row *sql.Row) (*models.Game, error) {
	g, err := scanGame(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return g, err
}

func (d *DB) CreateGame(ctx context.Context, creatorID int64, creatorName, gameType, initialBoard string) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx,
		`INSERT INTO games (creator_id, creator_name, game_type, board)
		VALUES ($1, $2, $3, $4)`+returning,
		creatorID, creatorName, gameType, initialBoard,
	)
	return scanGame(row)
}

func (d *DB) GetGame(ctx context.Context, id string) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx, selectAll+` WHERE id = $1`, id)
	return scanGameOrNil(row)
}

func (d *DB) GetGameByInlineMessageID(ctx context.Context, imid string) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx, selectAll+` WHERE inline_message_id = $1`, imid)
	return scanGameOrNil(row)
}

func (d *DB) SetInlineMessageID(ctx context.Context, gameID, imid string) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx,
		`UPDATE games SET inline_message_id = $2, updated_at = NOW() WHERE id = $1`+returning,
		gameID, imid,
	)
	return scanGameOrNil(row)
}

func (d *DB) JoinGame(ctx context.Context, gameID string, opponentID int64, opponentName string) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx,
		`UPDATE games SET opponent_id = $2, opponent_name = $3, status = 'ready', updated_at = NOW()
		WHERE id = $1 AND status = 'waiting' AND creator_id != $2 AND opponent_id IS NULL`+returning,
		gameID, opponentID, opponentName,
	)
	return scanGameOrNil(row)
}

func (d *DB) StartGame(ctx context.Context, gameID string, requesterID int64) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx,
		`UPDATE games SET status = 'playing', updated_at = NOW()
		WHERE id = $1 AND status = 'ready' AND creator_id = $2`+returning,
		gameID, requesterID,
	)
	return scanGameOrNil(row)
}

func (d *DB) UpdateBoard(ctx context.Context, gameID, newBoard, nextTurn string) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx,
		`UPDATE games SET board = $2, current_turn = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'playing'`+returning,
		gameID, newBoard, nextTurn,
	)
	return scanGameOrNil(row)
}

func (d *DB) FinishGame(ctx context.Context, gameID, result string) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx,
		`UPDATE games SET status = 'finished', result = $2, updated_at = NOW()
		WHERE id = $1`+returning,
		gameID, result,
	)
	return scanGameOrNil(row)
}

func (d *DB) QuitGame(ctx context.Context, gameID string, playerID int64) (*models.Game, error) {
	game, err := d.GetGame(ctx, gameID)
	if err != nil || game == nil || !game.IsPlayer(playerID) {
		return nil, err
	}

	var result string
	if game.Status == "waiting" {
		if playerID != game.CreatorID {
			return nil, nil
		}
		result = "cancelled"
	} else {
		if playerID == game.CreatorID {
			result = "forfeit_creator"
		} else {
			result = "forfeit_opponent"
		}
	}

	return d.FinishGame(ctx, gameID, result)
}

func (d *DB) CreateRematch(ctx context.Context, oldGame *models.Game, initialBoard string) (*models.Game, error) {
	newXPlayer := "opponent"
	if oldGame.XPlayer == "opponent" {
		newXPlayer = "creator"
	}

	_, err := d.conn.ExecContext(ctx,
		`UPDATE games SET inline_message_id = NULL WHERE id = $1`, oldGame.ID)
	if err != nil {
		return nil, err
	}

	row := d.conn.QueryRowContext(ctx,
		`INSERT INTO games (inline_message_id, game_type, creator_id, creator_name,
		opponent_id, opponent_name, status, x_player, board)
		VALUES ($1, $2, $3, $4, $5, $6, 'playing', $7, $8)`+returning,
		oldGame.InlineMessageID, oldGame.GameType, oldGame.CreatorID, oldGame.CreatorName,
		oldGame.OpponentID, oldGame.OpponentName, newXPlayer, initialBoard,
	)
	return scanGame(row)
}

func (d *DB) UpdateSelection(ctx context.Context, gameID string, selectedPos *int) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx,
		`UPDATE games SET selected_pos = $2, updated_at = NOW() WHERE id = $1`+returning,
		gameID, selectedPos,
	)
	return scanGameOrNil(row)
}

func (d *DB) UpdateBoardWithSelection(ctx context.Context, gameID, newBoard string, selectedPos *int) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx,
		`UPDATE games SET board = $2, selected_pos = $3, updated_at = NOW() WHERE id = $1`+returning,
		gameID, newBoard, selectedPos,
	)
	return scanGameOrNil(row)
}

func (d *DB) UpdateBoardClearSelection(ctx context.Context, gameID, newBoard, nextTurn string) (*models.Game, error) {
	row := d.conn.QueryRowContext(ctx,
		`UPDATE games SET board = $2, current_turn = $3, selected_pos = NULL, updated_at = NOW() WHERE id = $1`+returning,
		gameID, newBoard, nextTurn,
	)
	return scanGameOrNil(row)
}
