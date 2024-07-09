package parser

import (
	"context"
	"time"

	"github.com/gbandres98/wikidle2/internal/store"
)

type Parser struct {
	db *store.Queries
}

func New(db *store.Queries) *Parser {
	parser := &Parser{
		db: db,
	}

	err := parser.ParseArticle(context.TODO())
	if err != nil {
		panic(err)
	}

	return parser
}

func GetGameID(date time.Time) string {
	return date.Format("20060102")
}
