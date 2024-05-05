FROM mcr.microsoft.com/devcontainers/go:1-1.22-bookworm

# Postgresql client for debugging database
RUN apt-get update \
    && apt-get install -y postgresql-client

# Install `tailwindcss` and `just`
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss \
    && curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to /usr/local/bin

# Install `tern` as the IDE user to avoid mucking with permissions on various Go
# files.
USER vscode
RUN go install github.com/jackc/tern/v2@latest
