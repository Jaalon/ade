package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"automated_dev_environment/internal/orchestrator"
)

func main() {
	cfg := orchestrator.ConfigFromEnv()
	srv := orchestrator.NewServer(cfg)

	log.Printf("[ade-config] démarrage (version=%s)", orchestrator.Version)

	if err := srv.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "erreur: %v\n", err)
		os.Exit(1)
	}
}
