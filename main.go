package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func index(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger.Info("Handling index request.")

		io.WriteString(w, "Hello, world!")
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	logger.Info("Creating request handlers.")

	handler := http.NewServeMux()
	handler.HandleFunc("/", index(logger))

	server := http.Server{
		Addr:              "0.0.0.0:8000",
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	logger.Info("Starting server.", "address", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server failed.", "error", err)
	}
}
