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
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ctxKey int

const (
	idParamKey         ctxKey = 0
	paginationParamKey ctxKey = 1
	intervalParamKey   ctxKey = 2
	bodyParamKey       ctxKey = 3
)

var (
	ErrNotFound          = errors.New("not found")
	ErrBadRequest        = errors.New("bad request")
	ErrNotAZip           = errors.New("file is not a valid zip")
	ErrInvalidRoundOrder = errors.New("invalid round order")
	ErrDuplicateKey      = errors.New("duplicated key")
	ErrInvalidPlayerList = errors.New("invalid player list")
)

func setupRoutes(gc *GameController, rc *RoundController, tc *TurnController) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/games", func(r chi.Router) {
		//Get game
		r.With(IdInUrlParam).
			Get("/{id}", makeHTTPHandlerFunc(gc.findByID))

		// List games
		r.With(Pagination, Interval).
			Get("/", makeHTTPHandlerFunc(gc.list))

		// Create game
		r.With(
			middleware.AllowContentType("application/json"),
			JsonBody[CreateGameRequest]).
			Post("/", makeHTTPHandlerFunc(gc.create))

		// Update game
		r.With(IdInUrlParam,
			middleware.AllowContentType("application/json"),
			JsonBody[UpdateGameRequest]).
			Put("/{id}", makeHTTPHandlerFunc(gc.update))

		// Delete game
		r.With(IdInUrlParam).
			Delete("/{id}", makeHTTPHandlerFunc(gc.delete))

	})

	r.Route("/rounds", func(r chi.Router) {
		// Get round
		r.With(IdInUrlParam).
			Get("/{id}", makeHTTPHandlerFunc(rc.findByID))

		// List rounds
		r.Get("/", makeHTTPHandlerFunc(rc.list))

		// Create round
		r.With(middleware.AllowContentType("application/json"),
			JsonBody[CreateRoundRequest]).
			Post("/", makeHTTPHandlerFunc(rc.create))

		// Update round
		r.With(IdInUrlParam,
			middleware.AllowContentType("application/json"),
			JsonBody[UpdateRoundRequest]).
			Put("/{id}", makeHTTPHandlerFunc(rc.update))

		// Delete round
		r.With(IdInUrlParam).
			Delete("/{id}", makeHTTPHandlerFunc(rc.delete))

	})

	r.Route("/turns", func(r chi.Router) {
		// Get turn
		r.With(IdInUrlParam).
			Get("/{id}", makeHTTPHandlerFunc(tc.findByID))

		// List turn
		r.Get("/", makeHTTPHandlerFunc(tc.list))

		// Create turn
		r.With(middleware.AllowContentType("application/json"),
			JsonBody[CreateTurnsRequest]).
			Post("/", makeHTTPHandlerFunc(tc.create))

		// Update turn
		r.With(IdInUrlParam,
			middleware.AllowContentType("application/json"),
			JsonBody[UpdateTurnRequest]).
			Put("/{id}", makeHTTPHandlerFunc(tc.update))

		// Delete turn
		r.With(IdInUrlParam).
			Delete("/{id}", makeHTTPHandlerFunc(tc.delete))

		// Get turn file
		r.With(IdInUrlParam).
			Get("/{id}/files", makeHTTPHandlerFunc(tc.download))

		// Upload turn file
		r.With(IdInUrlParam,
			middleware.AllowContentType("application/zip")).
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

func Paginated(p *PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.page - 1) * p.pageSize
		return db.Offset(int(offset)).Limit(int(p.pageSize))
	}
}

func Intervaled(i *IntervalParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at between ? AND ?", i.startDate, i.endDate)
	}
}

func OrderBy(column string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderBy{
			Columns: []clause.OrderByColumn{
				{
					Column: clause.Column{
						Name: column,
					},
				},
			},
		})
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
		r = r.WithContext(context.WithValue(r.Context(), paginationParamKey, p))

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

		r = r.WithContext(context.WithValue(r.Context(), intervalParamKey, i))

		next.ServeHTTP(w, r)
		return nil
	}))
}

func IdInUrlParam(next http.Handler) http.Handler {
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

func JsonBody[T any](next http.Handler) http.Handler {
	return makeHTTPHandlerFunc((func(w http.ResponseWriter, r *http.Request) error {
		var t T

		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			return ApiError{
				code:    http.StatusBadRequest,
				Message: "Invalid json body",
			}
		}
		defer r.Body.Close()

		r = r.WithContext(context.WithValue(r.Context(), idParamKey, t))

		next.ServeHTTP(w, r)
		return nil
	}))
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

func handleError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrDuplicateKey
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
	var code int
	switch {
	case errors.Is(err, ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, ErrBadRequest),
		errors.Is(err, ErrInvalidPlayerList),
		errors.Is(err, ErrInvalidRoundOrder):
		code = http.StatusBadRequest
	case errors.Is(err, ErrNotAZip):
		code = http.StatusUnprocessableEntity
	case errors.Is(err, ErrDuplicateKey):
		code = http.StatusConflict
	default:
		code = http.StatusInternalServerError
	}

	return ApiError{code: code, Message: err.Error()}
}

type ApiFunction func(http.ResponseWriter, *http.Request) error

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

func mapToGameDTO(g *GameModel) *GameDto {
	return &GameDto{
		ID:           g.ID,
		CurrentRound: g.CurrentRound,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		Name:         g.Name,
		PlayersCount: g.PlayersCount,
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
