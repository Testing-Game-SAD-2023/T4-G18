package main

import (
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	mw "github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Configuration struct {
	PostgresUrl   string `json:"postgresUrl"`
	ListenAddress string `json:"listenAddress"`
	ApiPrefix     string `json:"apiPrefix"`
	DataDir       string `json:"dataDir"`
	BufferSize    int    `json:"bufferSize"`
	EnableSwagger bool   `json:"enableSwagger"`
}

//go:embed postman
var postmanDir embed.FS

func main() {
	var (
		configPath = flag.String("config", "config.json", "Path for configuration")
	)
	flag.Parse()

	fcontent, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	var configuration Configuration
	if err := json.Unmarshal(fcontent, &configuration); err != nil {
		log.Fatal(err)
	}

	makeDefaults(&configuration)

	if err := run(configuration); err != nil {
		log.Fatal(err)
	}
}

func run(c Configuration) error {

	db, err := gorm.Open(postgres.Open(c.PostgresUrl), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		return err
	}

	err = db.AutoMigrate(
		&GameModel{},
		&RoundModel{},
		&PlayerModel{},
		&TurnModel{},
		&MetadataModel{},
		&PlayerGameModel{})

	if err != nil {
		return err
	}

	if err := os.Mkdir(c.DataDir, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("cannot create data directory: %w", err)
	}

	r := chi.NewRouter()

	// serving Postman directory for documentation files
	fs := http.FileServer(http.FS(postmanDir))
	r.Mount("/public/", http.StripPrefix("/public/", fs))

	if c.EnableSwagger {
		r.Group(func(r chi.Router) {

			r.Use(cors.Handler(cors.Options{
				AllowedOrigins:   []string{"https://*", "http://*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
				ExposedHeaders:   []string{"Link"},
				AllowCredentials: false,
				MaxAge:           300, // Maximum value not ignored by any of major browsers
			}))

			opts := mw.SwaggerUIOpts{SpecURL: "/public/postman/schemas/index.yaml"}
			sh := mw.SwaggerUI(opts, nil)
			r.Handle("/docs", sh)
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
			})
		})
	}
	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		var (
			// game endpoint
			gameStorage    = NewGameStorage(db)
			gameService    = NewGameService(gameStorage)
			gameController = NewGameController(gameService)

			// round endpoint
			roundStorage    = NewRoundStorage(db)
			roundService    = NewRoundService(roundStorage)
			roundController = NewRoundController(roundService)

			// turn endpoint
			metadataStorage = NewMetadataStorage(db)
			turnStorage    = NewTurnStorage(db)
			turnService    = NewTurnService(turnStorage, metadataStorage, gameStorage, c.DataDir)
			turnController = NewTurnController(turnService, c.BufferSize)
		)

		r.Mount(c.ApiPrefix, setupRoutes(
			gameController,
			roundController,
			turnController,
		))
	})
	log.Printf("listening on %s", c.ListenAddress)
	return http.ListenAndServe(c.ListenAddress, r)

}

func makeDefaults(c *Configuration) {
	if c.ApiPrefix == "" {
		c.ApiPrefix = "/"
	}
	if c.ListenAddress == "" {
		c.ListenAddress = "localhost:3000"
	}

	if c.DataDir == "" {
		c.DataDir = "data"
	}
	if c.BufferSize <= 0 {
		c.BufferSize = 512
	}

}
