-- name: GetArticleByID :one
SELECT * FROM article 
WHERE id = $1;

-- name: SaveArticle :exec
INSERT INTO article (id, content, title)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET content = $2, title = $3;

-- name: SaveGame :exec
INSERT INTO game (player_id, game_id, game_data)
VALUES ($1, $2, $3)
ON CONFLICT (player_id, game_id) DO UPDATE SET game_data = $3;

-- name: GetQueueArticleByDate :one
SELECT * FROM article_queue
WHERE onDate = $1;

-- name: GetQueueArticle :one
SELECT * FROM article_queue
WHERE onDate IS NULL
ORDER BY id DESC
LIMIT 1;


-- name: AddArticleToQueue :exec
INSERT INTO article_queue (title)
VALUES ($1);

-- name: DeleteQueueArticle :exec
DELETE FROM article_queue
WHERE id = $1;

-- name: GetGameCountByGameID :one
SELECT COUNT(*) FROM game
WHERE game_id = $1;

-- name: GetWinCountByGameID :one
select COUNT(*) from game 
where game_id = $1 
and game_data->>'w' = 'true';

