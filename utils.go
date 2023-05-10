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

type ctxKey int

const (
	idParamKey          ctxKey = 0
	paginationParamsKey ctxKey = 1
	intervalParamsKey   ctxKey = 2
)

func setupRoutes(gc *GameController, rc *RoundController, tc *TurnController) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/games", func(r chi.Router) {
		//Get game
		r.With(ParamId).
			Get("/{id}", makeHTTPHandlerFunc(gc.findByID))

		// List games
		r.With(Pagination, Interval).
			Get("/", makeHTTPHandlerFunc(gc.list))

		// Create game
		r.With(ParamId, ContentType("application/json")).
			Post("/", makeHTTPHandlerFunc(gc.create))

		// Update game
		r.With(ParamId, ContentType("application/json")).
			Put("/{id}", makeHTTPHandlerFunc(gc.update))

		// Delete game
		r.With(ParamId).
			Delete("/{id}", makeHTTPHandlerFunc(gc.delete))

	})

	r.Route("/rounds", func(r chi.Router) {
		// Get round
		r.With(ParamId).
			Get("/{id}", makeHTTPHandlerFunc(rc.findByID))

		// List rounds
		r.Get("/", makeHTTPHandlerFunc(rc.list))

		// Create round
		r.With(ParamId, ContentType("application/json")).
			Post("/", makeHTTPHandlerFunc(rc.create))

		// Update round
		r.With(ParamId, ContentType("application/json")).
			Put("/{id}", makeHTTPHandlerFunc(rc.update))

		// Delete round
		r.With(ParamId).
			Delete("/{id}", makeHTTPHandlerFunc(rc.delete))

	})

	r.Route("/turns", func(r chi.Router) {
		// Get turn
		r.With(ParamId).
			Get("/{id}", makeHTTPHandlerFunc(tc.findByID))

		// List turn
		r.Get("/", makeHTTPHandlerFunc(tc.list))
		// Create turn
		r.With(ContentType("application/json")).
			Post("/", makeHTTPHandlerFunc(tc.create))

		// Update turn
		r.With(ParamId, ContentType("application/json")).
			Put("/{id}", makeHTTPHandlerFunc(tc.update))

		// Delete turn
		r.With(ParamId).
			Delete("/{id}", makeHTTPHandlerFunc(tc.delete))

		// Get turn file
		r.With(ParamId).
			Get("/{id}/files", makeHTTPHandlerFunc(tc.download))

		// Upload turn file
		r.With(ParamId, ContentType("application/zip")).
			Put("/{id}/files", makeHTTPHandlerFunc(tc.upload))
	})
	return r
}

type PaginationParams struct {
	page     int64
	pageSize int64
}

type IntervalParams struct {
	startDate time.Time
	endDate   time.Time
}

func PaginateScope(p *PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.page - 1) * p.pageSize
		return db.Offset(int(offset)).Limit(int(p.pageSize))
	}
}

func IntervalScope(i *IntervalParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at between ? AND ?", i.startDate, i.endDate)
	}
}

func Pagination(next http.Handler) http.Handler {
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
		r = r.WithContext(context.WithValue(r.Context(), paginationParamsKey, p))

		next.ServeHTTP(w, r)
		return nil
	}))
}

func Interval(next http.Handler) http.Handler {
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

		r = r.WithContext(context.WithValue(r.Context(), intervalParamsKey, i))

		next.ServeHTTP(w, r)
		return nil
	}))
}

func ParamId(next http.Handler) http.Handler {
	return makeHTTPHandlerFunc((func(w http.ResponseWriter, r *http.Request) error {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

		if err != nil {
			return ApiError{
				code:    http.StatusBadRequest,
				Message: "Invalid id",
			}
		}
		r = r.WithContext(context.WithValue(r.Context(), idParamKey, id))

		next.ServeHTTP(w, r)
		return nil
	}))
}
func ContentType(contentType string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cType := r.Header.Get("Content-Type")
			if cType != contentType {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
func parseNumberWithDefault(s string, d int64) (int64, error) {
	if s == "" {
		return d, nil
	}

	return strconv.ParseInt(s, 10, 64)
}

func parseDateWithDefault(s string, t time.Time) (time.Time, error) {
	if s == "" {
		return t, nil
	}
	return time.Parse("2006-01-02", s)

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

func handleDbError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	default:
		return err
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

func mapToGameDTO(g *GameModel) *GameDto {
	return &GameDto{
		ID:           g.ID,
		CurrentRound: g.CurrentRound,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		PlayersCount: g.PlayersCount,
		Name:         g.Name,
	}
}

func mapToRoundDTO(g *RoundModel) *RoundDto {
	return &RoundDto{
		ID:          g.ID,
		Order:       g.Order,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		TestClassId: g.TestClassId,
	}
}

func mapToTurnDTO(t *TurnModel) *TurnDto {
	return &TurnDto{
		ID:        t.ID,
		IsWinner:  t.IsWinner,
		Scores:    t.Scores,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		PlayerID:  t.PlayerID,
	}
}
