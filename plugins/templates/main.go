package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("[templates-plugin] ")

	plugin, err := NewTemplatePlugin()
	if err != nil {
		log.Fatalf("failed to create plugin: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Printf("shutting down...")
		cancel()
	}()

	if err := plugin.Start(ctx); err != nil {
		log.Fatalf("plugin error: %v", err)
	}
}
