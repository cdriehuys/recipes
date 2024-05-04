default:
    @just --list

clean:
    @rm -r ./static/*

dev:
    @mkdir static
    tailwindcss -i static-src/input.css -o static/style.css --watch
