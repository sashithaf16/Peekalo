package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sashithaf16/peekalo/config"
	"github.com/sashithaf16/peekalo/handler"
	"github.com/sashithaf16/peekalo/logger"
	"github.com/sashithaf16/peekalo/metrics"
)

func main() {

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{ // reference: https://go-chi.io/#/pages/middleware?id=cors

		AllowedOrigins: []string{"https://*", "http://*"},

		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	cfg := getConfig()
	logger := getLogger(getConfig())

	metrics.RegisterMetrics()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Application is healthy!"))
	})
	r.Handle("/metrics", promhttp.HandlerFor(metrics.PrometheusRegistry, promhttp.HandlerOpts{}))
	r.Post("/analyze", handler.NewAnalyzeUrlHandler(cfg, logger, http.DefaultClient).AnalyzeURLHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	logger.Info().Msg("Starting Peekalo server to analyze web pages...")

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Failed to start server")
			panic(err)
		}
	}()

	handleShutdown(srv, logger)
}

func handleShutdown(srv *http.Server, logger logger.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sig := <-stop
	logger.Info().Msgf("Shutdown signal received: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info().Msg("Starting graceful shutdown...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to shutdown server gracefully")
		panic(err)
	}
	logger.Info().Msg("Server shutdown gracefully")
}

func getConfig() *config.Config {
	config := config.GetConfig()
	return config
}

func getLogger(cfg *config.Config) logger.Logger {
	logger := logger.CreateLogger(cfg.LogLevel)
	return logger
}
