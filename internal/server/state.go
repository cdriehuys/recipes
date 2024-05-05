package server

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TemplateEngine interface {
	Write(w io.Writer, name string, data any) error
}

type State struct {
	Db             *pgxpool.Pool
	Logger         *slog.Logger
	TemplateEngine TemplateEngine
}

func (s *State) requestLogger(req *http.Request) *slog.Logger {
	logger := s.Logger.With(slog.Group("request", "method", req.Method, "path", req.URL.Path))
	logger.Info("Handling request.")

	return logger
}
