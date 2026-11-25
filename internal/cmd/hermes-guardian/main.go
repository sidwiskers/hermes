package main

import (
	"os"

	"github.com/sidwiskers/hermes/internal/guardian"
)

func main() {
	os.Exit(guardian.Execute(os.Args[1:]))
}
