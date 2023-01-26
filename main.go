package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rafaelespinoza/wrestic/internal/cmd"
)

var root cmd.Root

func init() {
	root = cmd.New()
}

func main() {
	if err := root.RunContext(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
