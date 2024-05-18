package migrations

import "embed"

//go:embed *.sql shared
var MigrationsFS embed.FS
