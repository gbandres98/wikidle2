package game

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gbandres98/wikidle2/internal/parser"
	"github.com/gbandres98/wikidle2/internal/templates"
)

type modalTemplateData struct {
	ArticleTitle string
	Attempts     int
	TotalPlayers int64
	TotalWins    int64
	Streak       int
	Words        []template.HTML
}

func (a *Api) createGameWinModalData(ctx context.Context, article parser.Article, playerData PlayerData) (modalTemplateData, error) {
	id := article.ID

	gameData, ok := playerData.Games[id]
	if !ok {
		return modalTemplateData{}, fmt.Errorf("game data not found for article %s", id)
	}

	totalPlayers, err := a.db.GetGameCountByGameID(ctx, id)
	if err != nil {
		return modalTemplateData{}, fmt.Errorf("failed to get games by game id: %w", err)
	}

	totalWins, err := a.db.GetWinCountByGameID(ctx, id)
	if err != nil {
		return modalTemplateData{}, fmt.Errorf("failed to get won games by game id: %w", err)
	}

	streak := 1

	idDate, err := time.Parse("20060102", id)
	if err != nil {
		return modalTemplateData{}, fmt.Errorf("failed to parse game id date: %w", err)
	}

	for {
		idDate = idDate.AddDate(0, 0, -1)
		nextID := parser.GetGameID(idDate)

		nextGameData, ok := playerData.Games[nextID]
		if !ok {
			break
		}

		if nextGameData.Won {
			streak++
		} else {
			break
		}
	}

	words := []template.HTML{}

	for i, word := range gameData.Words {
		words = append(words, template.HTML(fmt.Sprintf("<small style='display:block'>%d. %s</small>", i+1, word)))
	}

	return modalTemplateData{
		ArticleTitle: article.Title,
		Attempts:     len(gameData.Words),
		TotalPlayers: totalPlayers,
		TotalWins:    totalWins,
		Streak:       streak,
		Words:        words,
	}, nil
}

func (a *Api) writeGameWinModal(ctx context.Context, w http.ResponseWriter, article parser.Article, playerData PlayerData) error {
	tmplData, err := a.createGameWinModalData(ctx, article, playerData)
	if err != nil {
		return fmt.Errorf("failed to create game win modal data: %w", err)
	}

	return templates.Execute(w, "modal.html", tmplData)
}

func (a *Api) getGameWinModalContent(ctx context.Context, article parser.Article, playerData PlayerData) (template.HTML, error) {
	tmplData, err := a.createGameWinModalData(ctx, article, playerData)
	if err != nil {
		return "", fmt.Errorf("failed to create game win modal data: %w", err)
	}

	return templates.Render("modal.html", tmplData)
}
