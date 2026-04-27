package main

import (
	"log"
	"nocturne/scanner/internal/cli"
	"os"
)

func main() {
	// Initialize structured logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("🕯️ [NOCTURNE] ")

	// Check for debug mode via environment variable
	if os.Getenv("NOCTURNE_DEBUG") == "true" {
		log.Println("Debug mode enabled")
	}

	// Initialize and run the CLI
	app := cli.NewCLI()

	// Passing os.Args[1:] to skip the binary name
	app.Run(os.Args[1:])
}
