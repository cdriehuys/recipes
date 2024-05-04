build: clean
    @mkdir build
    @go build -o build/recipes .

test:
    go test ./...

clean:
    @rm -rf "./static/*" "./build"

dev:
    @mkdir static
    tailwindcss -i static-src/input.css -o static/style.css --watch
