// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package store

import (
	"encoding/json"
)

type Article struct {
	ID      string
	Content json.RawMessage
}

type Game struct {
	PlayerID string
	GameID   string
	GameData json.RawMessage
}