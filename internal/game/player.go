package game

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gbandres98/wikidle2/internal/parser"
	"github.com/gbandres98/wikidle2/internal/store"
	"github.com/google/uuid"
)

type GameData struct {
	Words     []string `json:"s"`
	Won       bool     `json:"w"`
	ArticleID string   `json:"i"`
}

type PlayerData struct {
	ID     string    `json:"i"`
	Game   *GameData `json:"g"`
	Streak int       `json:"s"`
}

func (a *Api) playerDataMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		articleID := parser.GetGameID(time.Now())
		playerData, err := readPlayerDataHeader(r, articleID)
		if err != nil {
			Error(w, err, 500, "", "error retrieving player data", articleID)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "playerData", playerData)
		ctx = context.WithValue(ctx, "articleID", articleID)

		next(w, r.WithContext(ctx))

		a.writePlayerDataHeader(w, playerData)
	}
}

func playerData(ctx context.Context) *PlayerData {
	return ctx.Value("playerData").(*PlayerData)
}

func articleID(ctx context.Context) string {
	return ctx.Value("articleID").(string)
}

func (a *Api) writePlayerDataHeader(w http.ResponseWriter, playerData *PlayerData) {
	go func() {
		err := a.storePlayerData(context.Background(), playerData)
		if err != nil {
			log.Printf("failed to store player data: %v\n", err)
		}
	}()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	err := json.NewEncoder(gz).Encode(playerData)
	if err != nil {
		log.Println(err)
		return
	}
	gz.Close()

	encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())
	log.Println(encodedString)
}

func readPlayerDataHeader(r *http.Request, articleID string) (*PlayerData, error) {
	playerData := &PlayerData{
		ID: uuid.NewString(),
		Game: &GameData{
			Words:     []string{},
			ArticleID: articleID,
		},
	}

	cookie, err := r.Cookie("gameData")
	if err != nil || cookie.Value == "" {
		return playerData, nil
	}

	err = json.Unmarshal([]byte(cookie.Value), playerData)
	if err != nil {
		log.Printf("error unmarshalling player data: %s", err)
		return playerData, nil
	}

	if playerData.Game.ArticleID != articleID {
		playerData.Game = &GameData{
			Words:     []string{},
			ArticleID: articleID,
		}
	}

	return playerData, nil
}

func (a *Api) storePlayerData(ctx context.Context, playerData *PlayerData) error {
	data, err := json.Marshal(playerData.Game)
	if err != nil {
		return err
	}

	err = a.db.SaveGame(ctx, store.SaveGameParams{
		PlayerID: playerData.ID,
		GameID:   playerData.Game.ArticleID,
		GameData: json.RawMessage(data),
	})
	if err != nil {
		return err
	}

	return nil
}
