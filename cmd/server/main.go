package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/archerny/mini-agent/internal/api"
	"github.com/archerny/mini-agent/internal/demo"
	"github.com/archerny/mini-agent/internal/runtime"
)

const (
	defaultAddr = ":8080"
)

func main() {
	fmt.Println("🚀 mini-agent — AI Agent Runtime")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println()

	// Create the runtime engine.
	engine := runtime.NewEngine()

	// Create HTTP router (REST + WebSocket).
	router := api.NewRouter(engine)

	// Determine listen address.
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = defaultAddr
	} else {
		addr = ":" + addr
	}

	// Create HTTP server.
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create a root context with signal handling.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT / SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n🛑 Received shutdown signal...")
		cancel()

		// Graceful HTTP shutdown.
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	// Start the demo scenario in the background.
	scenario := &demo.ResearchPipeline{}
	go func() {
		fmt.Printf("📡 Demo: %s\n", scenario.Name())
		if err := scenario.Run(ctx, engine); err != nil {
			log.Printf("[demo] scenario error: %v", err)
		}
	}()

	// Start HTTP server.
	fmt.Printf("🌐 HTTP server listening on %s\n", addr)
	fmt.Printf("   REST API:  http://localhost%s/api/agents\n", addr)
	fmt.Printf("   WebSocket: ws://localhost%s/ws/events\n", addr)
	fmt.Println()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}

	// Wait for graceful shutdown.
	fmt.Println("🛑 Shutting down agents...")
	engine.Shutdown(ctx)
	fmt.Println("✅ Server stopped.")
}
