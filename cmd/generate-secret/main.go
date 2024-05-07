package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/spf13/cobra"
)

func createCmd(w io.Writer, secretGen func(int) []byte) *cobra.Command {
	return &cobra.Command{
		Use:   "generate-secret",
		Short: "Generate secret suitable for secret or encryption key.",
		Long: "Generate random strings with the correct length and formatting for use as the " +
			"secret key or encryption key for the application.",
		Run: func(cmd *cobra.Command, args []string) {

			encoded := base64.StdEncoding.EncodeToString(secretGen(32))
			fmt.Fprintf(w, "%s\n", encoded)
		},
	}
}

func run(w io.Writer, secretGen func(int) []byte) error {
	cmd := createCmd(w, secretGen)

	return cmd.Execute()
}

func main() {
	if err := run(os.Stdout, securecookie.GenerateRandomKey); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
