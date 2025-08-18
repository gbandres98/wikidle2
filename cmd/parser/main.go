package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gbandres98/wikidle2/internal/parser"
	"github.com/gbandres98/wikidle2/internal/store"
	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v2"
)

var dbUrl, dbDriver, cronString, addr, forceTitle string
var force, show bool

func main() {
	app := &cli.App{
		Name:   "wikidle3-api",
		Usage:  "API server for wikidle3",
		Action: start,
		Commands: []*cli.Command{
			{
				Name:        "queue-category",
				Description: "Queue all articles from a wikipedia category id",
				Args:        true,
				ArgsUsage:   "Category id to process",
				Action:      queueCategory,
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
				},
			},
			{
				Name:        "force",
				Description: "Replace current article with a new one",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "title",
						Aliases:     []string{"t"},
						Usage:       "Article name when forcing a new article",
						Value:       "",
						Destination: &forceTitle,
					},
					&cli.BoolFlag{
						Name:        "show",
						Aliases:     []string{"s"},
						Usage:       "Show parsed article name when forcing a new article",
						Value:       false,
						Destination: &show,
					},
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
				},
				Action: replace,
			},
		},
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

func replace(c *cli.Context) error {
	ctx := c.Context

	db, err := store.NewDB(ctx, dbDriver, dbUrl, true)
	if err != nil {
		return err
	}

	p := parser.New(db)

	if forceTitle == "" {
		forceTitle, err = p.GetArticleTitleFromQueue(ctx)
		if err != nil {
			return err
		}
	}

	if show {
		log.Println(forceTitle)
	}
	return p.ParseArticle(ctx, forceTitle)
}

func queueCategory(c *cli.Context) error {
	ctx := c.Context

	db, err := store.NewDB(ctx, dbDriver, dbUrl, true)
	if err != nil {
		return err
	}

	type apiResponse struct {
		Continue struct {
			Code string `json:"cmcontinue"`
		} `json:"continue"`
		Query struct {
			Pages []struct {
				Title string `json:"title"`
			} `json:"categorymembers"`
		} `json:"query"`
	}

	req, err := url.Parse("https://es.wikipedia.org/w/api.php?format=json&formatversion=2&origin=*&action=query&list=categorymembers&uselang=content&cmtitle=" + url.QueryEscape("Categoría:Wikipedia:Artículos destacados") + "&cmlimit=500")
	if err != nil {
		panic(err)
	}

	pages := []string{}
	query := 1

	for {
		log.Printf("query %d", query)
		query++

		var response apiResponse

		res, err := http.Get(req.String())
		if err != nil {
			panic(err)
		}

		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			panic(err)
		}

		for _, page := range response.Query.Pages {
			pages = append(pages, page.Title)
		}

		if response.Continue.Code == "" {
			break
		}

		query := req.Query()
		query.Set("cmcontinue", response.Continue.Code)

		req.RawQuery = query.Encode()
	}

	log.Printf("got %d page names", len(pages))

	rand.Shuffle(len(pages), func(i, j int) { pages[i], pages[j] = pages[j], pages[i] })

	for i, page := range pages {
		log.Printf("storing page %d of %d", i, len(pages))
		err = db.AddArticleToQueue(ctx, page)
		if err != nil {
			panic(err)
		}
	}

	return nil
}
