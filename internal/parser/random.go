package parser

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type wikipediaRandomArticleResponseItem struct {
	Title string `json:"title"`
}

type wikipediaRandomArticleResponse struct {
	Items []wikipediaRandomArticleResponseItem `json:"items"`
}

func GetRandomArticleTitle() (string, error) {
	res, err := http.Get("https://es.wikipedia.org/api/rest_v1/page/random/title")
	if err != nil {
		return "", fmt.Errorf("failed to call random article api: %w", err)
	}
	defer res.Body.Close()

	var response wikipediaRandomArticleResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode random article response: %w", err)
	}

	if len(response.Items) < 1 {
		return "", fmt.Errorf("no items in random article response")
	}

	return response.Items[0].Title, nil
}

func (p *Parser) GetArticleTitleFromQueue(ctx context.Context) (string, error) {
	id := GetGameID(time.Now())

	article, err := p.db.GetQueueArticleByDate(ctx, sql.NullString{String: id, Valid: true})

	if err == sql.ErrNoRows {
		article, err = p.db.GetQueueArticle(ctx)
	}

	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to get article from queue: %w", err)
	}

	if err == nil {
		//err = p.db.DeleteQueueArticle(ctx, article.ID)
		if err != nil {
			return "", fmt.Errorf("failed to delete article from queue: %w", err)
		}

		return article.Title, nil
	}

	return GetRandomArticleTitle()
}
