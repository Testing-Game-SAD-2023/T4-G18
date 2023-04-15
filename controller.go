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
			apiError, ok := err.(ApiError)

			if ok {
				if err := writeJson(w, apiError.code, apiError); err != nil {
					log.Print(err)
				}
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			log.Print(err)
		}
	}

}
func MakeHTTPHandler(gc *GameService, rc *RoundService) *chi.Mux {
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

	rh := NewRoundHandler(rc)
	r.Route("/rounds", func(r chi.Router) {
		r.Get("/{id}", makeHTTPHandlerFunc(rh.create))

		r.Post("/", makeHTTPHandlerFunc(rh.findByID))

		r.Delete("/{id}", makeHTTPHandlerFunc(rh.delete))

		//r.Put

	})

	return r
}

type GameController struct {
	controller *GameService
}

func NewGameHandler(gc *GameService) *GameController {
	return &GameController{controller: gc}
}

func (gh *GameController) create(w http.ResponseWriter, r *http.Request) error {

	var request CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}

	defer r.Body.Close()

	g, err := gh.controller.Create(&request)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusCreated, gameModelToDto(g))

}

func (gh *GameController) findByID(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid game id",
		}
	}
	g, err := gh.controller.FindByID(id)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, gameModelToDto(g))

}

func (gh *GameController) delete(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid game id",
		}
	}

	if err := gh.controller.Delete(id); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type RoundController struct {
	controller *RoundService
}

func NewRoundHandler(rc *RoundService) *RoundController {
	return &RoundController{
		controller: rc,
	}
}

func (rh *RoundController) create(w http.ResponseWriter, r *http.Request) error {

	var request CreateRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
		}
	}

	defer r.Body.Close()

	g, err := rh.controller.Create(&request)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusCreated, roundModelToDto(g))

}

func (rh *RoundController) findByID(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid game id",
		}
	}
	round, err := rh.controller.FindByID(id)

	if err != nil {
		return makeApiError(err)
	}

	return writeJson(w, http.StatusOK, roundModelToDto(round))

}

func (rh *RoundController) delete(w http.ResponseWriter, r *http.Request) error {

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		return ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid round id",
		}
	}

	if err := rh.controller.Delete(id); err != nil {
		return makeApiError(err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
