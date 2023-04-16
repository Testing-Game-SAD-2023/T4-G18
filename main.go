package main

import (
	"encoding/json"
	"flag"
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
}

var (
	configPath = flag.String("config", "config.json", "Path for configuration")
)

func main() {
	flag.Parse()

	fcontent, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	var configuration Configuration
	if err := json.Unmarshal(fcontent, &configuration); err != nil {
		log.Fatal(err)
	}

	if err := run(configuration); err != nil {
		log.Fatal(err)
	}
}

func run(configuration Configuration) error {
	db, err := gorm.Open(postgres.Open(configuration.PostgresUrl), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&GameModel{}, &RoundModel{}); err != nil {
		return err
	}

	gameStorage := NewGameStorage(db)
	gameService := NewGameService(gameStorage)

	roundStorage := NewRoundStorage(db)
	roundService := NewRoundService(roundStorage)

	api := MakeHTTPHandler(gameService, roundService)

	r := chi.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Mount(configuration.ApiPrefix, api)
	})

	log.Printf("listening on %s", configuration.ListenAddress)
	return http.ListenAndServe(configuration.ListenAddress, r)

}
