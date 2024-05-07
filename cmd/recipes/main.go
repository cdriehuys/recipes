package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cdriehuys/recipes"
	"github.com/cdriehuys/recipes/internal/runtime"
)

func main() {
	err := runtime.Run(
		context.Background(),
		os.Stderr,
		os.Args[1:],
		recipes.StaticFS,
		recipes.TemplateFS,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}
