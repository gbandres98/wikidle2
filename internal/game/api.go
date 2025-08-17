package game

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gbandres98/wikidle2/internal/parser"
	"github.com/gbandres98/wikidle2/internal/store"
	"github.com/gbandres98/wikidle2/internal/templates"
)

type Api struct {
	db            *store.Queries
	baseAddress   string
	articleCache  bool
	cachedArticle parser.Article
}

func New(db *store.Queries, baseAddress string, articleCache bool) *Api {
	return &Api{
		db:           db,
		baseAddress:  baseAddress,
		articleCache: articleCache,
	}
}

func (a *Api) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /search", a.playerDataMiddleware(a.handleWordSearch))

	mux.HandleFunc("POST /init", a.playerDataMiddleware(a.handleInit))

	mux.HandleFunc("GET /{$}", a.handleGet)

}

func (a *Api) handleWordSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	article, err := a.getArticleOfTheDay(ctx, articleID(ctx))
	if err != nil {
		Error(w, err, http.StatusInternalServerError, "", "failed to get article of the day")
		return
	}

	playerData := playerData(ctx)

	newWord := r.FormValue("q")
	if len(strings.TrimSpace(newWord)) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	for _, word := range playerData.Game.Words {
		if parser.Normalize(word) == parser.Normalize(newWord) {
			w.WriteHeader(http.StatusAccepted)
			return
		}
	}

	for _, word := range parser.ExcludedWords {
		if word == parser.Normalize(newWord) {
			w.WriteHeader(http.StatusAccepted)
			return
		}
	}

	playerData.Game.Words = append(playerData.Game.Words, newWord)

	won := checkGameWin(playerData.Game, article)

	playerData.Game.Won = won

	if won {
		_, err := w.Write([]byte(fmt.Sprintf(`<div id="article" hx-swap-oob="true"><base href="//es.wikipedia.org/wiki/">%s</div>`, string(article.UnobscuredHTML))))
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write unobscured article")
			return
		}

		_, err = w.Write([]byte(fmt.Sprintf(`<p id="motd" hx-swap-oob="true">Adivinaste el artículo de hoy en %d intentos!</p>`, len(playerData.Game.Words))))
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write MOTD")
			return
		}

		err = a.writeGameWinModal(r.Context(), w, article, playerData)
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write game win modal")
			return
		}

		_, err = w.Write([]byte(`<script>onGameWin();</script>`))
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write onGameWin script")
			return
		}

		return
	}

	attIndex := len(playerData.Game.Words)

	hits, err := writeHits(w, newWord, attIndex, article)
	if err != nil {
		Error(w, err, http.StatusInternalServerError, "", "failed to write hits")
		return
	}

	_, err = w.Write([]byte(fmt.Sprintf(`<small onclick="scrollToNextWord(%d)">%d. %s - %d aciertos</small>`, attIndex, attIndex, r.FormValue("q"), hits)))
	if err != nil {
		Error(w, err, http.StatusInternalServerError, "", "failed to write hit word")
		return
	}

	err = a.writeClue(w, article, attIndex)
	if err != nil {
		Error(w, err, http.StatusInternalServerError, "", "failed to write clue")
		return
	}
}

func (a *Api) handleGet(w http.ResponseWriter, r *http.Request) {
	motd := "Adivina el artículo de hoy"

	modal := template.HTML("")

	article, err := a.getArticleOfTheDay(r.Context(), parser.GetGameID(time.Now()))
	if err != nil {
		Error(w, err, http.StatusInternalServerError, "", "failed to get article of the day")
		return
	}

	err = templates.Execute(w, "index.html", struct {
		BaseUrl           string
		Article           template.HTML
		Attempts          template.HTML
		MOTD              string
		Won               bool
		Modal             template.HTML
		SearchPlaceholder string
	}{
		BaseUrl:           a.baseAddress,
		Article:           article.HTML,
		Attempts:          template.HTML(""),
		MOTD:              motd,
		Won:               false,
		Modal:             modal,
		SearchPlaceholder: "Cargando el artículo de hoy...",
	})
	if err != nil {
		Error(w, err, http.StatusInternalServerError, "", "failed to execute template")
		return
	}
}

func (a *Api) handleInit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	article, err := a.getArticleOfTheDay(ctx, articleID(ctx))
	if err != nil {
		Error(w, err, http.StatusInternalServerError, "", "failed to get article of the day")
		return
	}

	playerData := playerData(ctx)

	if playerData.Game.Won {
		_, err := w.Write([]byte(fmt.Sprintf(`<div id="article" hx-swap-oob="true"><base href="//es.wikipedia.org/wiki/">%s</div>`, string(article.UnobscuredHTML))))
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write unobscured article")
			return
		}

		_, err = w.Write([]byte(fmt.Sprintf(`<p id="motd" hx-swap-oob="true">Adivinaste el artículo de hoy en %d intentos!</p>`, len(playerData.Game.Words))))
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write MOTD")
			return
		}

		err = a.writeGameWinModal(r.Context(), w, article, playerData)
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write game win modal")
			return
		}

		_, err = w.Write([]byte(`<script>onGameWin();</script>`))
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write onGameWin script")
			return
		}

		return
	}

	for attIndex, word := range playerData.Game.Words {
		hits, err := writeHits(w, word, attIndex, article)
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write hits")
			return
		}

		_, err = w.Write([]byte(fmt.Sprintf(`<small onclick="scrollToNextWord(%d)">%d. %s - %d aciertos</small>`, attIndex, attIndex, word, hits)))
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to write hit word")
			return
		}
	}

	err = a.writeClue(w, article, len(playerData.Game.Words))
	if err != nil {
		Error(w, err, http.StatusInternalServerError, "", "failed to write clue")
		return
	}
}

func Error(w http.ResponseWriter, err error, code int, message string, logMessage string, logArgs ...interface{}) {
	if message == "" {
		message = "Lo siento, ha ocurrido un error."
	}

	log.Printf(logMessage+"\n", logArgs...)
	log.Println(err)

	http.Error(w, message, code)
}
