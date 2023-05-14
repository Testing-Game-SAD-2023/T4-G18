package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxUploadSize   = 2 * (1 << 20) // 2MB
	defaultBodySize = 1 << 18       // 256KB

)

type ApiFunction func(http.ResponseWriter, *http.Request) error

func setupRoutes(gc *GameController, rc *RoundController, tc *TurnController, roc *RobotController) *chi.Mux {
	r := chi.NewRouter()

	r.Use(WithMaximumBodySize(defaultBodySize))

	r.Route("/games", func(r chi.Router) {
		//Get game
		r.Get("/{id}", makeHTTPHandlerFunc(gc.findByID))

		// List games
		r.Get("/", makeHTTPHandlerFunc(gc.list))

		// Create game
		r.With(middleware.AllowContentType("application/json")).
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
		r.With(middleware.AllowContentType("application/zip"),
			WithMaximumBodySize(maxUploadSize)).
			Put("/{id}/files", makeHTTPHandlerFunc(tc.upload))
	})

	r.Route("/robots", func(r chi.Router) {
		// Get robot with filter
		r.Get("/", makeHTTPHandlerFunc(roc.findByFilter))

		// Create robots in bulk
		r.With(middleware.AllowContentType("application/json")).
			Post("/", makeHTTPHandlerFunc(roc.createBulk))

		r.Delete("/", makeHTTPHandlerFunc(roc.delete))

	})

	return r
}

func WithPagination(p *PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.page - 1) * p.pageSize
		return db.Offset(int(offset)).Limit(int(p.pageSize))
	}
}

func WithInterval(i *IntervalParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at between ? AND ?", i.startDate, i.endDate)
	}
}

func WithOrder(column string) func(db *gorm.DB) *gorm.DB {
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

type Validable interface {
	Validate() error
}

func FromJsonBody[T Validable](r io.ReadCloser) (T, error) {

	var t T

	if err := json.NewDecoder(r).Decode(&t); err != nil {
		code := http.StatusBadRequest
		message := "Invalid json body"
		if err, ok := err.(*http.MaxBytesError); ok {
			code = http.StatusRequestEntityTooLarge
			message = fmt.Sprintf("allowed body size: %s", ByteCountIEC(err.Limit))
		}
		return t, ApiError{
			code:    code,
			Message: message,
			err:     err,
		}
	}
	defer r.Close()

	if err := t.Validate(); err != nil {
		return t, ApiError{
			code:    http.StatusBadRequest,
			Message: err.Error(),
			err:     err,
		}
	}

	return t, nil
}

type Convertable[T any] interface {
	Validable
	Convert(s string) (T, error)
}

func FromUrlParams[T Convertable[T]](r *http.Request, name string) (T, error) {
	s := chi.URLParam(r, name)
	return fromString[T](s, name)
}

func FromUrlQuery[T Convertable[T]](r *http.Request, name string, fallback T) (T, error) {
	s := r.URL.Query().Get(name)
	if s == "" {
		return fallback, nil
	}
	return fromString[T](s, name)
}

func fromString[T Convertable[T]](s, name string) (T, error) {
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
		return t, ApiError{
			code:    http.StatusBadRequest,
			err:     err,
			Message: err.Error(),
		}
	}

	return v, nil
}

func WithMaximumBodySize(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
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
					log.Print(apiError.err)
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
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
