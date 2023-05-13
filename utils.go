package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ApiFunction func(http.ResponseWriter, *http.Request) error

func setupRoutes(gc *GameController, rc *RoundController, tc *TurnController) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/games", func(r chi.Router) {
		//Get game
		r.Get("/{id}", makeHTTPHandlerFunc(gc.findByID))

		// List games
		r.Get("/", makeHTTPHandlerFunc(gc.list))

		// Create game
		r.With(
			middleware.AllowContentType("application/json"),
		).
			Post("/", makeHTTPHandlerFunc(gc.create))

		// Update game
		r.With(middleware.AllowContentType("application/json")).
			Put("/{id}", makeHTTPHandlerFunc(gc.update))

		// Delete game
		r.Delete("/{id}", makeHTTPHandlerFunc(gc.delete))

	})

	r.Route("/rounds", func(r chi.Router) {
		// Get round
		r.Get("/{id}", makeHTTPHandlerFunc(rc.findByID))

		// List rounds
		r.Get("/", makeHTTPHandlerFunc(rc.list))

		// Create round
		r.With(middleware.AllowContentType("application/json")).
			Post("/", makeHTTPHandlerFunc(rc.create))

		// Update round
		r.With(middleware.AllowContentType("application/json")).
			Put("/{id}", makeHTTPHandlerFunc(rc.update))

		// Delete round
		r.Delete("/{id}", makeHTTPHandlerFunc(rc.delete))

	})

	r.Route("/turns", func(r chi.Router) {
		// Get turn
		r.Get("/{id}", makeHTTPHandlerFunc(tc.findByID))

		// List turn
		r.Get("/", makeHTTPHandlerFunc(tc.list))

		// Create turn
		r.With(middleware.AllowContentType("application/json")).
			Post("/", makeHTTPHandlerFunc(tc.create))

		// Update turn
		r.With(middleware.AllowContentType("application/json")).
			Put("/{id}", makeHTTPHandlerFunc(tc.update))

		// Delete turn
		r.Delete("/{id}", makeHTTPHandlerFunc(tc.delete))

		// Get turn file
		r.Get("/{id}/files", makeHTTPHandlerFunc(tc.download))

		// Upload turn file
		r.With(middleware.AllowContentType("application/zip")).
			Put("/{id}/files", makeHTTPHandlerFunc(tc.upload))
	})
	return r
}

type Validator interface {
	Validate() error
}

func FromJsonBody[T Validator](r io.ReadCloser) (T, error) {

	var t T

	if err := json.NewDecoder(r).Decode(&t); err != nil {
		return t, ApiError{
			code:    http.StatusBadRequest,
			Message: "Invalid json body",
			err:     err,
		}
	}
	defer r.Close()

	if err := t.Validate(); err != nil {
		return t, err
	}

	return t, nil
}

type Convertable[T any] interface {
	Validator
	Convert(s string) (T, error)
}

func FromUrlParams[T Convertable[T]](r *http.Request, name string) (T, error) {
	s := chi.URLParam(r, name)
	var t T
	return fromString(s, name, t)
}

func FromUrlQuery[T Convertable[T]](r *http.Request, name string, fallback T) (T, error) {
	s := r.URL.Query().Get(name)
	return fromString(s, name, fallback)
}

func fromString[T Convertable[T]](s, name string, fallback T) (T, error) {
	var t T

	v, err := t.Convert(s)
	if err != nil {
		err = fmt.Errorf("%w %q: %v", ErrInvalidParam, name, err)
		return v, ApiError{
			code:    http.StatusBadRequest,
			err:     err,
			Message: err.Error(),
		}
	}

	if err := v.Validate(); err != nil {
		return v, err
	}

	return v, nil
}

func makePaginatedResponse(v any, count int64, p *PaginationParams) *PaginatedResponse {
	return &PaginatedResponse{
		Data: v,
		Metadata: PaginationMetadata{
			Count:    count,
			HasNext:  (count - p.page*p.pageSize) > 0,
			Page:     p.page,
			PageSize: p.pageSize,
		},
	}
}

func writeJson(w http.ResponseWriter, statusCode int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(v)
}

func makeHTTPHandlerFunc(f ApiFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			apiError, ok := err.(ApiError)

			if ok {
				if apiError.code == http.StatusInternalServerError {
					apiError.Message = "internal server error"
				}

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
