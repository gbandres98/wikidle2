package game

import (
	"fmt"
	"net/http"

	"github.com/gbandres98/wikidle2/internal/parser"
	"github.com/gbandres98/wikidle2/internal/templates"
)

func (a *Api) writeClue(w http.ResponseWriter, article parser.Article, attemptNumber int) error {
	if article.Clues == nil {
		return nil
	}

	if attemptNumber < 51 {
		return nil
	}

	clueMod := attemptNumber % 25

	var searchPlaceholder string
	if clueMod == 0 {
		clueIndex := (attemptNumber / 25) - 3

		if clueIndex >= len(article.Clues) {
			return nil
		}

		clue := article.Clues[clueIndex]

		searchPlaceholder = "Recibiste una pista!"

		_, err := w.Write([]byte(fmt.Sprintf(`<small>Pista %d: <strong>%s</strong></small>`, clueIndex+1, clue)))
		if err != nil {
			return err
		}
	} else {
		searchPlaceholder = fmt.Sprintf("Recibirás una pista en %d intentos", 25-clueMod)
	}

	search, err := templates.Render("search.html", struct {
		SearchPlaceholder string
	}{
		SearchPlaceholder: searchPlaceholder,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	_, err = w.Write([]byte(search))
	if err != nil {
		return err
	}

	return nil
}
