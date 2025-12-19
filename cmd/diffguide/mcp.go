package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mchowning/diffguide/internal/mcpserver"
	"github.com/mchowning/diffguide/internal/storage"
)

func runMCP(verbose bool) {
	store, err := storage.NewStore()
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	srv := mcpserver.New(store)

	ctx, cancel := context.WithCancel(context.Background())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		if verbose {
			log.Println("Shutting down MCP server...")
		}
		cancel()
	}()

	if verbose {
		log.Println("Starting MCP server on stdio...")
	}

	if err := srv.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("MCP server error: %v", err)
	}
}
