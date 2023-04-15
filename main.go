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

	var configuration Configuration
	fcontent, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

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

	err = db.AutoMigrate(&GameModel{}, &RoundModel{})
	if err != nil {
		return err
	}

	r := chi.NewRouter()

	gameStorage := NewGameStorage(db)
	gameController := NewGameController(gameStorage)

	roundStorage := NewRoundStorage(db)
	roundController := NewRoundController(roundStorage)

	api := MakeHTTPHandler(gameController, roundController)

	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Mount(configuration.ApiPrefix, api)
	})

	log.Printf("listening on %s", configuration.ListenAddress)
	return http.ListenAndServe(configuration.ListenAddress, r)

}
