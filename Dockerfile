FROM debian:stable AS css-build

WORKDIR /srv/recipes

# Install the `tailwindcss` CLI
RUN mkdir static \
    && apt-get update \
    && apt-get install -y curl \
    && curl -sL https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.3/tailwindcss-linux-x64 -o ./tailwindcss \
    && chmod +x tailwindcss

COPY tailwind.config.js .
COPY static-src ./static-src
COPY templates ./templates
RUN ./tailwindcss -i static-src/input.css -o static/style.css --minify


FROM golang:1.22-bookworm AS builder

WORKDIR /opt/recipes

# Pre-download modules
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY --from=css-build /srv/recipes/static ./static
COPY migrations ./migrations
COPY templates ./templates
COPY internal ./internal
COPY main.go ./

RUN go build -o build/recipes

# TODO: Move this to a scratch or minimalist image
ENTRYPOINT [ "/opt/recipes/build/recipes" ]
