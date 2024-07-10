package parser

import (
	"time"

	"github.com/gbandres98/wikidle2/internal/store"
)

type Parser struct {
	db *store.Queries
}

func New(db *store.Queries) *Parser {
	return &Parser{
		db: db,
	}
}

func GetGameID(date time.Time) string {
	return date.Format("20060102")
}
