package game

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gbandres98/wikidle2/internal/parser"
	"github.com/gbandres98/wikidle2/internal/store"
	"github.com/gbandres98/wikidle2/internal/templates"
	"github.com/google/uuid"
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
	mux.HandleFunc("POST /search", func(w http.ResponseWriter, r *http.Request) {
		article, err := a.getArticleOfTheDay(r.Context())
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to get article of the day")
			return
		}

		playerData, err := readCookie(r)
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to read cookie")
			return
		}

		gameData, ok := playerData.Games[article.ID]
		if !ok {
			gameData = GameData{
				Words:   []string{},
				Article: article.Title,
			}
		}

		newWord := r.FormValue("q")
		if len(strings.TrimSpace(newWord)) == 0 {
			w.WriteHeader(http.StatusAccepted)
			return
		}

		for _, word := range gameData.Words {
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

		gameData.Words = append(gameData.Words, newWord)

		won := checkGameWin(gameData, article)

		gameData.Won = won

		playerData.Games[article.ID] = gameData
		err = a.saveCookie(w, playerData)
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to save cookie: %+v", playerData)
			return
		}

		if won {
			_, err := w.Write([]byte(fmt.Sprintf(`<div id="article" hx-swap-oob="true"><base href="//es.wikipedia.org/wiki/">%s</div>`, string(article.UnobscuredHTML))))
			if err != nil {
				Error(w, err, http.StatusInternalServerError, "", "failed to write unobscured article")
				return
			}

			_, err = w.Write([]byte(fmt.Sprintf(`<p id="motd" hx-swap-oob="true">Adivinaste el artículo de hoy en %d intentos!</p>`, len(gameData.Words))))
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

		attIndex := len(gameData.Words)

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
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		article, err := a.getArticleOfTheDay(r.Context())
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to get article of the day")
			return
		}

		playerData, err := readCookie(r)
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to read cookie")
			return
		}

		gameData, ok := playerData.Games[article.ID]
		if !ok {
			gameData = GameData{
				Words: []string{},
			}
		}

		won := checkGameWin(gameData, article)

		newArticle, attempts, err := a.init(gameData, article)
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to init game")
			return
		}

		motd := "Adivina el artículo de hoy"

		modal := template.HTML("")

		if won {
			newArticle = article.UnobscuredHTML
			motd = fmt.Sprintf("Adivinaste el artículo de hoy en %d intentos!", len(gameData.Words))
			modal, err = a.getGameWinModalContent(r.Context(), article, playerData)
			if err != nil {
				Error(w, err, http.StatusInternalServerError, "", "failed to get game win modal content")
				return
			}
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
			Article:           newArticle,
			Attempts:          attempts,
			MOTD:              motd,
			Won:               won,
			Modal:             modal,
			SearchPlaceholder: "Prueba una palabra...",
		})
		if err != nil {
			Error(w, err, http.StatusInternalServerError, "", "failed to execute template")
			return
		}
	})
}

func (a *Api) saveCookie(w http.ResponseWriter, playerData PlayerData) error {
	go func() {
		err := a.storePlayerData(context.Background(), playerData)
		if err != nil {
			log.Printf("failed to store player data: %v\n", err)
		}
	}()

	todayGameID := parser.GetGameID(time.Now())

	for gameID, game := range playerData.Games {
		if gameID == todayGameID {
			continue
		}

		game.Words = []string{}

		playerData.Games[gameID] = game
	}

	dataBytes, err := json.Marshal(playerData)
	if err != nil {
		return err
	}

	base64Data := base64.StdEncoding.EncodeToString(dataBytes)

	w.Header().Add("HX-Trigger", fmt.Sprintf(`{ "afterResponse": "%s" }`, base64Data))

	return nil
}

func readCookie(r *http.Request) (PlayerData, error) {
	playerData := PlayerData{
		ID:    uuid.New().String(),
		Games: map[string]GameData{},
	}

	cookie := r.Header.Get("gameData")
	if cookie == "" {
		return playerData, nil
	}

	dataBytes, err := base64.StdEncoding.DecodeString(cookie)
	if err != nil {
		return playerData, err
	}

	err = json.Unmarshal(dataBytes, &playerData)
	if err != nil {
		return playerData, err
	}

	return playerData, nil
}

func (a *Api) storePlayerData(ctx context.Context, playerData PlayerData) error {
	gameID := parser.GetGameID(time.Now())

	game, ok := playerData.Games[gameID]
	if !ok {
		return nil
	}

	data, err := json.Marshal(game)
	if err != nil {
		return err
	}

	err = a.db.SaveGame(ctx, store.SaveGameParams{
		PlayerID: playerData.ID,
		GameID:   gameID,
		GameData: json.RawMessage(data),
	})
	if err != nil {
		return err
	}

	return nil
}

func Error(w http.ResponseWriter, err error, code int, message string, logMessage string, logArgs ...interface{}) {
	if message == "" {
		message = "Lo siento, ha ocurrido un error."
	}

	log.Printf(logMessage+"\n", logArgs...)
	log.Println(err)

	http.Error(w, message, code)
}
