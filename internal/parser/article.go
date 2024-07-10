package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/dlclark/regexp2"
	"github.com/gbandres98/wikidle2/internal/store"
)

var tokenizerRegex = regexp2.MustCompile(`(?<=>[^<>]*)(\b[\w]+\b)(?=[^<>]*<)`, regexp2.None)

type Article struct {
	ID    string
	Title string
	// Normalized word -> list of obscured spans
	Tokens      map[string][]int
	TitleTokens []string
	// Obscured span id -> original word
	Words          map[int]string
	HTML           template.HTML
	UnobscuredHTML template.HTML
}

func (p *Parser) ParseArticle(ctx context.Context, articleTitle string) error {
	article := Article{
		ID:          GetGameID(time.Now()),
		Title:       articleTitle,
		Tokens:      make(map[string][]int),
		TitleTokens: make([]string, 0),
		Words:       make(map[int]string),
	}

	article.Title = strings.Replace(article.Title, "_", " ", -1)

	res, err := http.Get("https://es.wikipedia.org/api/rest_v1/page/html/" + url.PathEscape(article.Title))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	newBody, err := tokenizerRegex.Replace(string(bodyBytes), `<span class="obscured">$1</span>`, -1, -1)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(newBody))
	if err != nil {
		return err
	}

	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		s.RemoveAttr("data-mw")
		s.RemoveAttr("id")
	})

	doc.Find("audio").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	doc.Find(".vcard").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	doc.Find(".mw-authority-control").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	title := `<h1>`
	titleWords := strings.Split(article.Title, " ")
	for _, word := range titleWords {
		title += `<span class="obscured">` + word + `</span> `
		article.TitleTokens = append(article.TitleTokens, Normalize(word))
	}
	title += `</h1>`

	doc.Find("section").First().BeforeHtml(title)

	unobscuredHTML, err := doc.Html()
	if err != nil {
		return err
	}

	article.UnobscuredHTML = template.HTML(unobscuredHTML)

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		s.SetAttr("href", "javascript:void(0);")
		s.RemoveAttr("title")
	})

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		height := s.AttrOr("height", "400")
		width := s.AttrOr("width", "400")

		s.SetAttr("src", "https://placehold.co/"+width+"x"+height+"?text=?")
		s.RemoveAttr("srcset")
	})

	doc.Find("span.obscured").Each(func(i int, s *goquery.Selection) {
		word := Normalize(s.Text())

		if IsExcludedWord(word) {
			return
		}

		if _, ok := article.Tokens[word]; !ok {
			article.Tokens[word] = []int{}
		}
		article.Tokens[word] = append(article.Tokens[word], i)
		article.Words[i] = s.Text()

		s.SetAttr("id", "obscured-"+fmt.Sprint(i))
		s.SetText(strings.Repeat("#", utf8.RuneCountInString(s.Text())))
	})

	html, err := doc.Html()
	if err != nil {
		return err
	}

	article.HTML = template.HTML(html)

	articleJson, err := json.Marshal(article)
	if err != nil {
		return err
	}

	return p.db.SaveArticle(ctx, store.SaveArticleParams{
		ID:      article.ID,
		Content: articleJson,
		Title:   article.Title,
	})
}
