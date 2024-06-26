recipes_bin := justfile_directory() / "cmd/recipes"

# Compile a production-ready application build
build: build-css
    go build -o build/recipes "{{ recipes_bin }}"

# Build the CSS for the application
build-css:
    tailwindcss -i static-src/input.css -o static/style.css --minify

# Run application tests
test *OPTS='':
    @# Embedding the static directory requires it to exist and have embeddable
    @# files.
    @mkdir -p static
    @touch static/style.css
    go test {{OPTS}} ./...

# Test with coverage reporting
test-cov: (test '-race' '-coverprofile=coverage.out' '-covermode=atomic')
    go tool cover -html=coverage.out -o=coverage.html

# Remove all generated artifacts
clean:
    @rm -rfv ./build ./static/* __debug_bin*

# Watch and recompile web assets for development
dev:
    @mkdir -p static
    tailwindcss -i static-src/input.css -o static/style.css --watch

# Open shell connected to dev database
db-shell:
    @psql --username {{ env_var('POSTGRES_USER') }} --host {{ env_var('POSTGRES_HOSTNAME') }}

migration_dir := justfile_directory() / "migrations"

# Migrate the database to the latest version
migrate: (_tern "migrate")

# Migration targets may be a migration number, a positive or negative delta, or
# 0 to revert all migrations.
#
# Migrate to a particular state
migrate-to target: (_tern "migrate" "--destination" target)

# Create a new migration
new-migration name: (_tern "new" name)

# Use `tern` to execute migrations from the correct working directory.
_tern +ARGS:
    #!/usr/bin/env bash
    set -eufo pipefail
    cd {{migration_dir}}
    tern {{ARGS}}
