package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

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

type ApiError struct {
	Message string `json:"message"`
}

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

	api := setupRoutes(gameController, roundController)

	r.Use(middleware.Logger)
	r.Mount("/", api)

	http.ListenAndServe(configuration.ListenAddress, r)

}

func setupRoutes(gc *GameController, rc *RoundController) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/games", func(r chi.Router) {
		// Create Game
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var request CreateGameRequest

			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			defer r.Body.Close()

			g, err := gc.Create(&request)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			writeJson(w, http.StatusCreated, gameModelToDto(g))
		})

		//Get Game
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			game, err := gc.FindByID(id)
			if errors.Is(err, ErrNotFound) {
				writeJson(w, http.StatusNotFound, ApiError{
					Message: "Resource not found",
				})
				return
			}

			writeJson(w, http.StatusOK, gameModelToDto(game))

		})

		// sus
		// r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
		// 	//gameId, _ := strconv.Atoi(chi.URLParam(r, "id"))

		// })

		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			gameId, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			err = gc.Delete(gameId)
			if errors.Is(err, ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			} else if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})

	r.Route("/rounds", func(r chi.Router) {
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {

		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {

		})

		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {

		})

		//r.Put

	})

	return r
}

func gameModelToDto(g *GameModel) *GameDto {
	return &GameDto{
		ID:           g.ID,
		CurrentRound: g.CurrentRound,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		PlayersCount: g.PlayersCount,
	}
}

func roundModelToDto(g *RoundModel) *RoundDto {
	return &RoundDto{
		ID:          g.ID,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		IdTestClass: g.IdTestClass,
	}
}

func writeJson(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
