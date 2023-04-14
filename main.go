package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Configuration struct {
	PostgresUrl   string `json:"postgresUrl"`
	ListenAddress string `json:"listenAddress"`
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
	json.Unmarshal(fcontent, &configuration)

	db, err := gorm.Open(postgres.Open(configuration.PostgresUrl), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&GameModel{}, &RoundModel{})
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	gameStorage := NewGameStorage(db)
	gameController := NewGameController(gameStorage)

	roundStorage := NewRoundStorage(db)
	roundController := NewRoundController(roundStorage)

	api := MakeHTTPHandler(gameController, roundController)

	r.Use(middleware.Logger)
	r.Mount("/", api)

	http.ListenAndServe(configuration.ListenAddress, r)

}


