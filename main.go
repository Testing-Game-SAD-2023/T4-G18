package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	mw "github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Configuration struct {
	PostgresUrl     string        `json:"postgresUrl"`
	ListenAddress   string        `json:"listenAddress"`
	ApiPrefix       string        `json:"apiPrefix"`
	DataDir         string        `json:"dataDir"`
	EnableSwagger   bool          `json:"enableSwagger"`
	CleanupInterval time.Duration `json:"cleanupInterval"`
}

//go:embed postman
var postmanDir embed.FS

func main() {
	var (
		configPath = flag.String("config", "config.json", "Path for configuration")
	)
	flag.Parse()

	fcontent, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	var configuration Configuration
	if err := json.Unmarshal(fcontent, &configuration); err != nil {
		log.Fatal(err)
	}

	makeDefaults(&configuration)

	if err := run(configuration); err != nil {
		log.Fatal(err)
	}
}

func run(c Configuration) error {

	db, err := gorm.Open(postgres.Open(c.PostgresUrl), &gorm.Config{
		SkipDefaultTransaction: true,
		TranslateError:         true,
	})

	if err != nil {
		return err
	}

	err = db.AutoMigrate(
		&GameModel{},
		&RoundModel{},
		&PlayerModel{},
		&TurnModel{},
		&MetadataModel{},
		&PlayerGameModel{})

	if err != nil {
		return err
	}

	if err := os.Mkdir(c.DataDir, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("cannot create data directory: %w", err)
	}

	r := chi.NewRouter()

	// basic cors
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept", "Authorization"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	if c.EnableSwagger {
		r.Group(func(r chi.Router) {
			opts := mw.SwaggerUIOpts{SpecURL: "/public/postman/schemas/index.yaml"}
			sh := mw.SwaggerUI(opts, nil)
			r.Handle("/docs", sh)
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
			})
		})
	}

	// serving Postman directory for documentation files
	fs := http.FileServer(http.FS(postmanDir))
	r.Mount("/public/", http.StripPrefix("/public/", fs))

	// metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		var (

			// game endpoint
			gameRepository = NewGameRepository(db)
			gameController = NewGameController(gameRepository)

			// round endpoint
			roundRepository = NewRoundStorage(db)
			roundController = NewRoundController(roundRepository)

			// turn endpoint
			turnRepository = NewTurnRepository(db, c.DataDir)
			turnController = NewTurnController(turnRepository)
		)

		r.Mount(c.ApiPrefix, setupRoutes(
			gameController,
			roundController,
			turnController,
		))
	})
	log.Printf("listening on %s", c.ListenAddress)

	done := make(chan struct{})
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)

	go func() {
		defer close(done)
		log.Printf("received: %s", <-signals)
	}()

	go func() {
		ticker := time.NewTicker(c.CleanupInterval)
		for {
			select {
			case <-done:
			case <-ticker.C:
				log.Print("cleanup")
			}

		}
	}()
	return startHttpServer(r, c.ListenAddress, done)

}

func startHttpServer(r chi.Router, addr string, done <-chan struct{}) error {
	server := http.Server{
		Addr:              addr,
		Handler:           r,
		ReadTimeout:       time.Minute,
		WriteTimeout:      time.Minute,
		IdleTimeout:       time.Hour,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error)
	defer close(errCh)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-done:
	case err := <-errCh:
		return err
	}

	ctx, canc := context.WithTimeout(context.Background(), time.Second*10)
	defer canc()

	return server.Shutdown(ctx)
}

func makeDefaults(c *Configuration) {
	if c.ApiPrefix == "" {
		c.ApiPrefix = "/"
	}

	if c.ListenAddress == "" {
		c.ListenAddress = "localhost:3000"
	}

	if c.DataDir == "" {
		c.DataDir = "data"
	}

	if int64(c.CleanupInterval) == 0 {
		c.CleanupInterval = time.Hour
	}

}
