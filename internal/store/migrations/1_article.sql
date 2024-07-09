-- +goose Up
CREATE TABLE article (
    id VARCHAR(8) NOT NULL PRIMARY KEY,
    content JSONB NOT NULL
);