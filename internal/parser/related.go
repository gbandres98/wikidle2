package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type relatedResponse struct {
	Pages []struct {
		Title  string `json:"title"`
		Titles struct {
			Canonical  string `json:"canonical"`
			Normalized string `json:"normalized"`
		} `json:"titles"`
	} `json:"pages"`
}

func (p *Parser) parseRelated(ctx context.Context, articleTitle string) ([]string, error) {
	res, err := http.Get(fmt.Sprintf("https://es.wikipedia.org/api/rest_v1/page/related/%s", articleTitle))
	if err != nil {
		return nil, fmt.Errorf("error getting related articles: %w", err)
	}

	var related relatedResponse
	if err := json.NewDecoder(res.Body).Decode(&related); err != nil {
		return nil, fmt.Errorf("error decoding related articles: %w", err)
	}

	var relatedTitles []string
	for _, page := range related.Pages {
		relatedTitles = append(relatedTitles, page.Titles.Normalized)
	}

	return relatedTitles, nil
}
