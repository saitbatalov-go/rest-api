package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"rest-api/internal/config"
	"rest-api/internal/http-server/handlers/url/redirect"
	"rest-api/internal/http-server/handlers/url/save"
	"rest-api/internal/lib/logger/handlers/slogpretty"
	"rest-api/internal/lib/logger/sl"
	"rest-api/internal/storage/sqlite"

	mwLogger "rest-api/internal/http-server/middleware/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

func main() {
	fmt.Print("11111111")
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info(
		"starting server rest-api",
		slog.String("env", cfg.Env),
		slog.String("version", "123"),
	)
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("rest-api", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("start server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Failed to start server", sl.Err(err))
	}

	log.Info("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = setupPrettySlog()
	case envProd:
		log = setupPrettySlog()
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
	}
	handler := opts.NewPrettyHandler(os.Stdout)
	return slog.New(handler)
}
