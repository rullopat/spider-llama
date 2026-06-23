package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rullopat/spider-llama/internal/config"
	"github.com/rullopat/spider-llama/internal/provider"
	"github.com/rullopat/spider-llama/internal/router"
	"github.com/rullopat/spider-llama/internal/server"
)

func main() {
	configPath := flag.String("config", "spider-llama.json", "path to spider-llama JSON config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	registry := router.NewRegistry(cfg)
	providers := provider.NewRegistry()
	app := server.New(cfg, registry, providers)

	httpServer := &http.Server{
		Addr:              cfg.Listen,
		Handler:           app.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		log.Printf("spider-llama listening on %s", cfg.Listen)
		errs <- httpServer.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errs:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Fatalf("shutdown failed: %v", err)
		}
	}
}
