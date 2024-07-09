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
	db          *store.Queries
	baseAddress string
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		playerData, err := readCookie(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gameData, ok := playerData.Games[article.ID]
		if !ok {
			gameData = GameData{
				Words: []string{},
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

		gameData.Words = append(gameData.Words, newWord)

		won := checkGameWin(gameData, article)

		gameData.Won = won

		playerData.Games[article.ID] = gameData
		err = a.saveCookie(w, playerData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if won {
			_, err := w.Write([]byte(fmt.Sprintf(`<div id="article" hx-swap-oob="true">%s</div>`, string(article.UnobscuredHTML))))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = w.Write([]byte(fmt.Sprintf(`<p id="motd" hx-swap-oob="true">Adivinaste el artículo de hoy en %d intentos!</p>`, len(gameData.Words))))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = w.Write([]byte(`<script>onGameWin();</script>`))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			return
		}

		hits := 0

		if indexes, ok := article.Tokens[parser.Normalize(newWord)]; ok {
			for _, i := range indexes {
				word, ok := article.Words[i]
				if !ok {
					continue
				}

				_, err = w.Write([]byte(fmt.Sprintf(`<span id="obscured-%d" hx-swap-oob="true" class="hit">%s</span>`, i, word)))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				hits++
			}
		}

		_, err = w.Write([]byte(fmt.Sprintf(`<small>%s - %d aciertos</small>`, r.FormValue("q"), hits)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		article, err := a.getArticleOfTheDay(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		playerData, err := readCookie(r)
		if err != nil {
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

		if won {
			newArticle = article.UnobscuredHTML
			motd = fmt.Sprintf("Adivinaste el artículo de hoy en %d intentos!", len(gameData.Words))
		}

		err = templates.Execute(w, "index.html", struct {
			BaseUrl  string
			Article  template.HTML
			Attempts template.HTML
			MOTD     string
			Won      bool
		}{
			BaseUrl:  a.baseAddress,
			Article:  newArticle,
			Attempts: attempts,
			MOTD:     motd,
			Won:      won,
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
		MaxAge:   0,
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

func (a *Api) getArticleOfTheDay(ctx context.Context) (parser.Article, error) {
	gameID := parser.GetGameID(time.Now())

	storeArticle, err := a.db.GetArticleByID(ctx, gameID)
	if err != nil {
		return parser.Article{}, err
	}

	var article parser.Article
	err = json.Unmarshal(storeArticle.Content, &article)
	if err != nil {
		return parser.Article{}, err
	}

	return article, nil
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
