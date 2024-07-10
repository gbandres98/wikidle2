-- +goose Up
CREATE TABLE article_queue (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    onDate VARCHAR(8)
);