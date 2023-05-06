package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func gameModelToDto(g *GameModel) *GameDto {
	return &GameDto{
		ID:           g.ID,
		CurrentRound: g.CurrentRound,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		PlayersCount: g.PlayersCount,
		Name:         g.Name,
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

func turnModelToDto(t *TurnModel) *TurnDto {
	return &TurnDto{
		ID:          t.ID,	
		IsWinner:  	 t.IsWinner,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,	
		PlayerID:    t.PlayerID,
	}
}

func writeJson(w http.ResponseWriter, statusCode int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(v)
}

func makeApiError(err error) error {

	switch {
	case errors.Is(err, ErrNotFound):
		return ApiError{code: http.StatusNotFound, Message: "Resource not found"}
	case errors.Is(err, ErrBadRequest):
		return ApiError{code: http.StatusBadRequest, Message: "Bad request"}
	case errors.Is(err, ErrNotAZip):
		return ApiError{code: http.StatusUnprocessableEntity, Message: "File is not a valid zip"}
	default:
		return ApiError{code: http.StatusInternalServerError, Message: "Internal server error"}
	}
}

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

func setupRoutes(gc *GameController, rc *RoundController, tc *TurnController) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/games", func(r chi.Router) {
		// Create Game
		r.Post("/", makeHTTPHandlerFunc(gc.create))

		//Get Game
		r.Get("/{id}", makeHTTPHandlerFunc(gc.findByID))

		// r.Put
		r.Put("/{id}", makeHTTPHandlerFunc(gc.update))

		r.Delete("/{id}", makeHTTPHandlerFunc(gc.delete))
	})

	r.Route("/rounds", func(r chi.Router) {
		r.Get("/{id}", makeHTTPHandlerFunc(rc.findByID))

		r.Post("/", makeHTTPHandlerFunc(rc.create))

		r.Delete("/{id}", makeHTTPHandlerFunc(rc.delete))

		//r.Put

	})

	r.Route("/turns", func(r chi.Router) {
		r.Get("/{id}", makeHTTPHandlerFunc(tc.findByID))
		r.Post("/", makeHTTPHandlerFunc(tc.create))
		r.Delete("/{id}", makeHTTPHandlerFunc(tc.delete))

		r.Put("/{id}/files", makeHTTPHandlerFunc(tc.upload))
		r.Get("/{id}/files", makeHTTPHandlerFunc(tc.download))
	})
	return r
}
