-- +goose Up
ALTER TABLE article ADD COLUMN title TEXT NOT NULL;