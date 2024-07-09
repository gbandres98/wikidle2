-- name: GetArticleByID :one
SELECT * FROM article 
WHERE id = $1;

-- name: SaveArticle :exec
INSERT INTO article (id, content)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET content = $2;

-- name: SaveGame :exec
INSERT INTO game (player_id, game_id, game_data)
VALUES ($1, $2, $3)
ON CONFLICT (player_id, game_id) DO UPDATE SET game_data = $3;