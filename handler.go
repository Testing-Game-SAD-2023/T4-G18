package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ApiFunction func(http.ResponseWriter, *http.Request) error

func makeHTTPHandlerFunc(f ApiFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			log.Print(err)
		}

	}
}

func MakeHTTPHandler(gc *GameController, rc *RoundController) *chi.Mux {
	r := chi.NewRouter()

	gh := NewGameHandler(gc)
	r.Route("/games", func(r chi.Router) {
		// Create Game
		r.Post("/", makeHTTPHandlerFunc(gh.create))

		//Get Game
		r.Get("/{id}", makeHTTPHandlerFunc(gh.findByID))

		// sus
		// r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
		// 	//gameId, _ := strconv.Atoi(chi.URLParam(r, "id"))

		// })

		r.Delete("/{id}", makeHTTPHandlerFunc(gh.delete))
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

type GameHandler struct {
	controller *GameController
}

func NewGameHandler(gc *GameController) *GameHandler {
	return &GameHandler{controller: gc}
}

func (gh *GameHandler) create(w http.ResponseWriter, r *http.Request) error {

	var request CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Cannot parse json body",
		}
	}

	defer r.Body.Close()

	g, err := gh.controller.Create(&request)

	if err != nil {
		return err
	}

	return writeJson(w, http.StatusCreated, gameModelToDto(g))

}

func (gh *GameHandler) findByID(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid game id",
		}
	}
	g, err := gh.controller.FindByID(id)

	if err != nil {
		return err
	}

	return writeJson(w, http.StatusCreated, gameModelToDto(g))

}

func (gh *GameHandler) delete(w http.ResponseWriter, r *http.Request) error {

	gameId, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid game id",
		}
	}

	if err := gh.controller.Delete(gameId); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type RoundHandler struct {
	controller *RoundController
}
