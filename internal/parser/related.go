package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type relatedResponse struct {
	Query struct {
		Pages []struct {
			Title string `json:"title"`
		} `json:"pages"`
	} `json:"query"`
}

func (p *Parser) parseRelated(ctx context.Context, articleTitle string) ([]string, error) {
	url := fmt.Sprintf("https://es.wikipedia.org/w/api.php?format=json&formatversion=2&origin=*&action=query&generator=search&gsrnamespace=0&gsrlimit=10&gsrqiprofile=classic&uselang=content&gsrsearch=morelike:%s", articleTitle)

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting related articles: %w", err)
	}

	var related relatedResponse
	if err := json.NewDecoder(res.Body).Decode(&related); err != nil {
		return nil, fmt.Errorf("error decoding related articles: %w", err)
	}

	var relatedTitles []string
	for _, page := range related.Query.Pages {
		relatedTitles = append(relatedTitles, page.Title)
	}

	return relatedTitles, nil
}
