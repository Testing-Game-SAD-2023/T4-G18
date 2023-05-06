package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
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
		ID:        t.ID,
		IsWinner:  t.IsWinner,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		PlayerID:  t.PlayerID,
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
		// Create game
		r.Post("/", makeHTTPHandlerFunc(gc.create))

		//Get game
		r.Get("/{id}", makeHTTPHandlerFunc(gc.findByID))

		// Update game
		r.Put("/{id}", makeHTTPHandlerFunc(gc.update))

		// Delete game
		r.Delete("/{id}", makeHTTPHandlerFunc(gc.delete))

		r.With(WithPagination, WithInterval).Get("/", makeHTTPHandlerFunc(gc.list))
	})

	r.Route("/rounds", func(r chi.Router) {
		r.Get("/{id}", makeHTTPHandlerFunc(rc.findByID))
		r.Post("/", makeHTTPHandlerFunc(rc.create))
		r.Delete("/{id}", makeHTTPHandlerFunc(rc.delete))
		r.Get("/", makeHTTPHandlerFunc(rc.list))

	})

	r.Route("/turns", func(r chi.Router) {
		r.Get("/{id}", makeHTTPHandlerFunc(tc.findByID))
		r.Get("/", makeHTTPHandlerFunc(tc.list))
		r.Post("/", makeHTTPHandlerFunc(tc.create))
		r.Delete("/{id}", makeHTTPHandlerFunc(tc.delete))
		r.Put("/{id}/files", makeHTTPHandlerFunc(tc.upload))
		r.Get("/{id}/files", makeHTTPHandlerFunc(tc.download))
	})
	return r
}

type PaginationParams struct {
	page     int
	pageSize int
}

type IntervalParams struct {
	startDate time.Time
	endDate   time.Time
}

func PaginateScope(p *PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		offset := (p.page - 1) * p.pageSize
		return db.Offset(offset).Limit(p.pageSize)
	}
}

func IntervalScope(i *IntervalParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		return db.Where("created_at between ? AND ?", i.startDate, i.endDate)
	}
}

func WithPagination(next http.Handler) http.Handler {
	return makeHTTPHandlerFunc((func(w http.ResponseWriter, r *http.Request) error {
		q := r.URL.Query()

		page, err := parseNumberWithDefault(q.Get("page"), 1)
		if err != nil {
			return writeJson(w, http.StatusBadRequest, ApiError{Message: "invalid 'page' parameter"})
		}

		if page <= 0 {
			page = 1
		}

		pageSize, err := parseNumberWithDefault(q.Get("pageSize"), 10)
		if err != nil {
			return writeJson(w, http.StatusBadRequest, ApiError{Message: "invalid 'pageSize'parameter"})
		}
		switch {
		case pageSize >= 100:
			pageSize = 100

		case pageSize <= 0:
			pageSize = 10
		}

		p := PaginationParams{
			pageSize: pageSize,
			page:     page,
		}
		r = r.WithContext(context.WithValue(r.Context(), "paginationParams", p))

		next.ServeHTTP(w, r)
		return nil
	}))
}

func WithInterval(next http.Handler) http.Handler {
	return makeHTTPHandlerFunc((func(w http.ResponseWriter, r *http.Request) error {
		q := r.URL.Query()

		startDate, err := parseDateWithDefault(q.Get("startDate"), time.Now().Add(-24*time.Hour))

		if err != nil {
			return writeJson(w, http.StatusBadRequest, ApiError{Message: "invalid parameter 'startDate'"})
		}

		endDate, err := parseDateWithDefault(q.Get("endDate"), time.Now())

		if err != nil {
			return writeJson(w, http.StatusBadRequest, ApiError{Message: "invalid parameter 'endDate'"})
		}

		i := IntervalParams{
			startDate: startDate,
			endDate:   endDate,
		}

		r = r.WithContext(context.WithValue(r.Context(), "intervalParams", i))

		next.ServeHTTP(w, r)
		return nil
	}))
}

func parseNumberWithDefault(s string, d int) (int, error) {
	if s == "" {
		return d, nil
	}

	return strconv.Atoi(s)
}

func parseDateWithDefault(s string, t time.Time) (time.Time, error) {
	if s == "" {
		return t, nil
	}
	return time.Parse("2006-01-02", s)

}

func makePaginatedResponse(v any, count int, p *PaginationParams) *PaginatedResponse {
	return &PaginatedResponse{
		Data: v,
		Metadata: PaginationMetadata{
			Count:   count,
			HasNext: (count - p.page * p.pageSize) > 0,
		},
	}
}
