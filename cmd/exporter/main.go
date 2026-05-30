package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ragtech-supervise/internal/api"
	"ragtech-supervise/internal/collector"
	"ragtech-supervise/internal/metrics"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ragtechURL := getEnv("RAGTECH_URL", "http://localhost:4470")
	listenAddr := getEnv("EXPORTER_ADDR", ":4471")

	client := api.NewClient(ragtechURL)
	reg := metrics.NewRegistry()

	upsCollector := collector.NewUPSCollector(client)
	reg.MustRegister(upsCollector)

	http.HandleFunc("/health", healthHandler(client))
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	logger.Info("starting exporter", "addr", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func healthHandler(client *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := client.Ping(ctx); err != nil {
			http.Error(w, "Ragtech API unreachable: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
