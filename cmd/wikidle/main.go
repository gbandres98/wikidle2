package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gbandres98/wikidle2/internal/game"
	"github.com/gbandres98/wikidle2/internal/static"
	"github.com/gbandres98/wikidle2/internal/store"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli/v2"
)

var dbUrl, dbDriver, addr, baseAddress string
var articleCache bool

func main() {
	app := &cli.App{
		Name:   "wikidle3-api",
		Usage:  "API server for wikidle3",
		Action: start,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "db-driver",
				EnvVars:     []string{"WIKIDLE_DATABASE_DRIVER"},
				Usage:       "Database driver to use (sqlite3, postgres)",
				Value:       "sqlite3",
				Destination: &dbDriver,
			},
			&cli.StringFlag{
				Name:        "db-dsn",
				EnvVars:     []string{"WIKIDLE_DATABASE_DSN"},
				Usage:       "Database connection string",
				Value:       "file::memory:?cache=shared",
				Destination: &dbUrl,
			},
			&cli.StringFlag{
				Name:        "listen-addr",
				EnvVars:     []string{"WIKIDLE_LISTEN_ADDRESS"},
				Usage:       "Address to listen on",
				Value:       "0.0.0.0:8080",
				Destination: &addr,
			},
			&cli.StringFlag{
				Name:        "base-addr",
				EnvVars:     []string{"WIKIDLE_BASE_ADDRESS"},
				Usage:       "Base address for web content",
				Value:       "http://192.168.1.100:8080",
				Destination: &baseAddress,
			},
			&cli.BoolFlag{
				Name:        "article-cache",
				EnvVars:     []string{"WIKIDLE_ARTICLE_CACHE"},
				Usage:       "Use article cache",
				Value:       true,
				Destination: &articleCache,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Panic(err)
	}
}

func start(c *cli.Context) error {
	ctx := c.Context

	db, err := store.NewDB(ctx, dbDriver, dbUrl, true)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	mux.Handle("GET /{file...}", http.FileServerFS(static.FS()))

	game := game.New(db, baseAddress, articleCache)
	game.RegisterHandlers(mux)

	log.Printf("Server started at %s\n", addr)
	return http.ListenAndServe(addr, mux)
}
