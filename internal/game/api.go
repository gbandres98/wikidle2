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

	"github.com/gbandres98/wikidle2/internal/parser"
	"github.com/gbandres98/wikidle2/internal/store"
	"github.com/gbandres98/wikidle2/internal/templates"
	"github.com/google/uuid"
)

type Api struct {
	db            *store.Queries
	baseAddress   string
	cachedArticle parser.Article
}

func New(db *store.Queries, baseAddress string) *Api {
	return &Api{
		db:          db,
		baseAddress: baseAddress,
	}
}

func (a *Api) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /search", func(w http.ResponseWriter, r *http.Request) {
		article, err := a.getArticleOfTheDay(r.Context())
		if err != nil {
			log.Printf("36 - %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		playerData, err := readCookie(r)
		if err != nil {
			log.Printf("42 - %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			log.Printf("54 - empty word\n")
			w.WriteHeader(http.StatusAccepted)
			return
		}

		for _, word := range gameData.Words {
			if parser.Normalize(word) == parser.Normalize(newWord) {
				log.Printf("61 - repeated word %+v %+v\n", word, gameData.Words)
				w.WriteHeader(http.StatusAccepted)
				return
			}
		}

		for _, word := range parser.ExcludedWords {
			if word == parser.Normalize(newWord) {
				log.Printf("69 - excluded word %+v\n", newWord)
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
			log.Printf("82 - %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if won {
			_, err := w.Write([]byte(fmt.Sprintf(`<div id="article" hx-swap-oob="true">%s</div>`, string(article.UnobscuredHTML))))
			if err != nil {
				log.Printf("90 - %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = w.Write([]byte(fmt.Sprintf(`<p id="motd" hx-swap-oob="true">Adivinaste el artículo de hoy en %d intentos!</p>`, len(gameData.Words))))
			if err != nil {
				log.Printf("96 - %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = a.writeGameWinModal(r.Context(), w, article, playerData)
			if err != nil {
				log.Printf("103 - %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = w.Write([]byte(`<script>onGameWin();</script>`))
			if err != nil {
				log.Printf("110 - %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		}

		attIndex := len(gameData.Words)

		hits, err := writeHits(w, newWord, attIndex, article)
		if err != nil {
			log.Printf("120 - %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(fmt.Sprintf(`<small onclick="scrollToNextWord(%d)">%d. %s - %d aciertos</small>`, attIndex, attIndex, r.FormValue("q"), hits)))
		if err != nil {
			log.Printf("127 - %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = a.writeClue(w, article, attIndex)
		if err != nil {
			log.Printf("134 - %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		article, err := a.getArticleOfTheDay(r.Context())
		if err != nil {
			log.Printf("142 - %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		playerData, err := readCookie(r)
		if err != nil {
			log.Printf("148 - %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		motd := "Adivina el artículo de hoy"

		modal := template.HTML("")

		if won {
			newArticle = article.UnobscuredHTML
			motd = fmt.Sprintf("Adivinaste el artículo de hoy en %d intentos!", len(gameData.Words))
			modal, err = a.getGameWinModalContent(r.Context(), article, playerData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	dataBytes, err := json.Marshal(playerData)
	if err != nil {
		return err
	}

	base64Data := base64.StdEncoding.EncodeToString(dataBytes)

	cookie := http.Cookie{
		Name:     "playerData",
		Value:    base64Data,
		Path:     "/",
		MaxAge:   31536000,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	return nil
}

func readCookie(r *http.Request) (PlayerData, error) {
	playerData := PlayerData{
		ID:    uuid.New().String(),
		Games: map[string]GameData{},
	}

	cookie, err := r.Cookie("playerData")
	if err != nil {
		return playerData, nil
	}

	dataBytes, err := base64.StdEncoding.DecodeString(cookie.Value)
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
	for gameID, game := range playerData.Games {
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
	}

	return nil
}
