// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: queries.sql

package store

import (
	"context"
	"encoding/json"
)

const getArticleByID = `-- name: GetArticleByID :one
SELECT id, content FROM article 
WHERE id = $1
`

func (q *Queries) GetArticleByID(ctx context.Context, id string) (Article, error) {
	row := q.db.QueryRowContext(ctx, getArticleByID, id)
	var i Article
	err := row.Scan(&i.ID, &i.Content)
	return i, err
}

const saveArticle = `-- name: SaveArticle :exec
INSERT INTO article (id, content)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET content = $2
`

type SaveArticleParams struct {
	ID      string
	Content json.RawMessage
}

func (q *Queries) SaveArticle(ctx context.Context, arg SaveArticleParams) error {
	_, err := q.db.ExecContext(ctx, saveArticle, arg.ID, arg.Content)
	return err
}

const saveGame = `-- name: SaveGame :exec
INSERT INTO game (player_id, game_id, game_data)
VALUES ($1, $2, $3)
ON CONFLICT (player_id, game_id) DO UPDATE SET game_data = $3
`

type SaveGameParams struct {
	PlayerID string
	GameID   string
	GameData json.RawMessage
}

func (q *Queries) SaveGame(ctx context.Context, arg SaveGameParams) error {
	_, err := q.db.ExecContext(ctx, saveGame, arg.PlayerID, arg.GameID, arg.GameData)
	return err
}
