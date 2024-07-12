package game

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gbandres98/wikidle2/internal/parser"
)

type GameData struct {
	Words   []string
	Won     bool
	Article string
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

	for i, word := range gameData.Words {
		hits := 0

		if indexes, ok := article.Tokens[parser.Normalize(word)]; ok {
			wordNumber := 0
			for _, index := range indexes {
				doc.Find(fmt.Sprintf("#obscured-%d", index)).First().SetText(article.Words[index]).AddClass(fmt.Sprintf("word-%d-%d", i+1, wordNumber))
				hits++
				wordNumber++
			}
		}

		attemptsHtml += fmt.Sprintf(`<small onclick="scrollToNextWord(%d)">%d. %s - %d aciertos</small>`, i+1, i+1, word, hits)
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
			if parser.Normalize(word) == titleToken || parser.IsExcludedWord(titleToken) {
				remaining--
				break
			}
		}
	}

	return remaining == 0
}

func (a *Api) getArticleOfTheDay(ctx context.Context) (parser.Article, error) {
	gameID := parser.GetGameID(time.Now())

	if a.cachedArticle.ID == gameID {
		return a.cachedArticle, nil
	}

	storeArticle, err := a.db.GetArticleByID(ctx, gameID)
	if err != nil {
		return parser.Article{}, err
	}

	var article parser.Article
	err = json.Unmarshal(storeArticle.Content, &article)
	if err != nil {
		return parser.Article{}, err
	}

	a.cachedArticle = article

	return article, nil
}

func writeHits(w http.ResponseWriter, word string, attIndex int, article parser.Article) (int, error) {
	hits := 0

	if indexes, ok := article.Tokens[parser.Normalize(word)]; ok {
		for n, i := range indexes {
			word, ok := article.Words[i]
			if !ok {
				continue
			}

			_, err := w.Write([]byte(fmt.Sprintf(`<span id="obscured-%d" hx-swap-oob="true" class="hit word-%d-%d">%s</span>`, i, attIndex, n, word)))
			if err != nil {
				return 0, err
			}

			hits++
		}
	}

	return hits, nil
}
