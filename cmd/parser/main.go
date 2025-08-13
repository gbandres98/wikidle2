package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gbandres98/wikidle2/internal/parser"
	"github.com/gbandres98/wikidle2/internal/store"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
)

var dbUrl, dbDriver, cronString, addr, forceTitle string
var force bool

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
				Name:        "cron",
				EnvVars:     []string{"WIKIDLE_PARSER_CRON"},
				Usage:       "Cron to run the article parsing job",
				Value:       "0 0 * * *",
				Destination: &cronString,
			},
			&cli.StringFlag{
				Name:        "listen-addr",
				EnvVars:     []string{"WIKIDLE_LISTEN_ADDRESS"},
				Usage:       "Address to listen on",
				Value:       "0.0.0.0:8080",
				Destination: &addr,
			},
			&cli.BoolFlag{
				Name:        "force",
				Usage:       "Force a new article",
				Value:       false,
				Destination: &force,
			},
			&cli.StringFlag{
				Name:        "force-title",
				Usage:       "Article name when forcing a new article",
				Value:       "",
				Destination: &forceTitle,
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

	p := parser.New(db)

	if force {
		if forceTitle == "" {
			forceTitle, err = p.GetArticleTitleFromQueue(ctx)
			if err != nil {
				return err
			}
		}

		return p.ParseArticle(ctx, forceTitle)
	}

	_, err = db.GetArticleByID(ctx, parser.GetGameID(time.Now()))
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("Error getting retrieving article from db: %v", err)
	}

	if err == sql.ErrNoRows {
		log.Println("No article in db, parsing article from queue")
		articleTitle, err := p.GetArticleTitleFromQueue(ctx)
		if err != nil {
			return err
		}

		err = p.ParseArticle(ctx, articleTitle)
		if err != nil {
			return err
		}
	}

	cr := cron.New()

	_, err = cr.AddFunc(cronString, func() {
		log.Printf("Running article parsing job at %v\n", time.Now())

		gameID := parser.GetGameID(time.Now())

		_, err := db.GetArticleByID(ctx, gameID)
		if err != nil && err != sql.ErrNoRows {
			panic(fmt.Errorf("Error getting retrieving article from db: %v", err))
		}

		if err == nil {
			log.Printf("Article already exists in db for game id %s\n", gameID)
			return
		}

		log.Printf("No article in db for game id %s, parsing article from queue\n", gameID)
		articleTitle, err := p.GetArticleTitleFromQueue(ctx)
		if err != nil {
			panic(err)
		}

		err = p.ParseArticle(ctx, articleTitle)
		if err != nil {
			panic(err)
		}
	})
	if err != nil {
		return err
	}

	go func() {
		mux := http.DefaultServeMux
		mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {})

		err := http.ListenAndServe(addr, mux)
		log.Println(err)
	}()

	log.Printf("Started cron schedule with %s\n", cronString)
	cr.Run()

	return nil
}
