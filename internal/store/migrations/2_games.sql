-- +goose Up
CREATE TABLE game (
    player_id TEXT NOT NULL,
    game_id VARCHAR(8) NOT NULL,
    game_data JSONB NOT NULL,
    PRIMARY KEY (player_id, game_id)
);