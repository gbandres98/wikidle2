package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type relatedResponse struct {
	Query struct {
		Pages []struct {
			Categories []struct {
				Title string `json:"title"`
			} `json:"categories"`
		} `json:"pages"`
	} `json:"query"`
}

func (p *Parser) parseRelated(articleTitle string) ([]string, error) {
	url := fmt.Sprintf("https://es.wikipedia.org/w/api.php?format=json&formatversion=2&origin=*&action=query&prop=categories&titles=%s", url.PathEscape(articleTitle))

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting related articles: %w", err)
	}

	var related relatedResponse
	if err := json.NewDecoder(res.Body).Decode(&related); err != nil {
		log.Println(url)
		return nil, fmt.Errorf("error decoding related articles: %w", err)
	}

	var relatedTitles []string
	for _, category := range related.Query.Pages[0].Categories {
		titleSplit := strings.Split(category.Title, ":")
		title := titleSplit[len(titleSplit)-1]

		relatedTitles = append(relatedTitles, title)
	}

	return relatedTitles, nil
}
