package game

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gbandres98/wikidle2/internal/parser"
)

type GameData struct {
	Words []string
	Won   bool
}

type PlayerData struct {
	ID    string
	Games map[string]GameData
}

func (a *Api) init(gameData GameData, article parser.Article) (newArticle template.HTML, attempts template.HTML, err error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(article.HTML)))
	if err != nil {
		return "", "", err
	}

	attemptsHtml := ""

	for _, word := range gameData.Words {
		hits := 0

		if indexes, ok := article.Tokens[parser.Normalize(word)]; ok {
			for _, index := range indexes {
				doc.Find(fmt.Sprintf("#obscured-%d", index)).First().SetText(article.Words[index])
				hits++
			}
		}

		attemptsHtml += fmt.Sprintf(`<small>%s - %d aciertos</small>`, word, hits)
	}

	articleHtml, err := doc.Html()
	if err != nil {
		return "", "", err
	}
	newArticle = template.HTML(articleHtml)

	attempts = template.HTML(attemptsHtml)

	return
}

func checkGameWin(gameData GameData, article parser.Article) bool {
	remaining := len(article.TitleTokens)

	for _, titleToken := range article.TitleTokens {
		for _, word := range gameData.Words {
			if parser.Normalize(word) == titleToken {
				remaining--
				break
			}
		}
	}

	return remaining == 0
}
