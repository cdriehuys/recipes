package server

import (
	"io"
	"log/slog"

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
