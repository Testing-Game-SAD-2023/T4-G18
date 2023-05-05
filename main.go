package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
}

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

	if err := validateAndMakeDefaults(&configuration); err != nil {
		log.Fatal(err)
	}

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

	if err := db.AutoMigrate(&GameModel{}, &RoundModel{}, &PlayerModel{}, &TurnModel{}, &MetadataModel{}, &PlayerGameModel{}); err != nil {
		return err
	}

	if err := os.Mkdir(c.DataDir, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("cannot create data directory: %w", err)
	}

	r := chi.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(middleware.Logger)

		// custom middleware to allow only json and multipart data in POST, PUT, PATCH requests
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				cType := r.Header.Get("Content-Type")
				switch r.Method {
				case http.MethodPost, http.MethodPatch:
					if cType != "application/json" {
						w.WriteHeader(http.StatusUnsupportedMediaType)
						return
					}
				case http.MethodPut:
					if cType != "application/json" && cType != "application/zip" {
						w.WriteHeader(http.StatusUnsupportedMediaType)
						return
					}
				}
				next.ServeHTTP(w, r)
			})
		})
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
			turnStorage    = NewTurnStorage(db)
			turnService    = NewTurnService(turnStorage, c.DataDir)
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

func validateAndMakeDefaults(c *Configuration) error {
	if c.ApiPrefix == "" {
		c.ApiPrefix = "/"
	}
	if c.ListenAddress == "" {
		c.ListenAddress = "localhost:3000"
	}

	if c.DataDir == "" {
		c.DataDir = "data"
	}
	if c.BufferSize == 0 {
		c.BufferSize = 512
	} else if c.BufferSize < 0 {
		return fmt.Errorf("buffer size must be a positive integer")
	}

	return nil

}
