package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lucchmielowski/kyverno-agent/internal/logger"
	"github.com/lucchmielowski/kyverno-agent/pkg/tools"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	port       int
	stdio      bool
	kubeconfig *string

	Name = "kyverno-tool-server"

	// Set for production with ldflags
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "kyverno-tool-server",
	Short: "Kyverno tool server",
	Run:   run,
}

func init() {
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	rootCmd.Flags().BoolVarP(&stdio, "stdio", "s", false, "Use stdio for communication instead of HTTP")
	kubeconfig = rootCmd.Flags().String("kubeconfig", "k", "Path to kubeconfig (optional, defaults to in-cluster config)")
}

func run(cmd *cobra.Command, args []string) {
	// TODO: Add logger
	logger.Init(stdio)
	defer logger.Sync()

	// Ctx setup for raceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Get().Info("Starting "+Name, "version", Version, "git_commit", GitCommit, "build_date", BuildDate)

	mcp := server.NewMCPServer(Name, Version)

	// Register kyverno tool to MCP server
	tools.RegisterTools(mcp)

	var wg sync.WaitGroup
	// Signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	var httpServer *http.Server

	wg.Add(1)
	if stdio {
		go func() {
			defer wg.Done()
			runStdioServer(ctx, mcp)
		}()
	} else {
		sseServer := server.NewStreamableHTTPServer(mcp, server.WithHeartbeatInterval(30*time.Second))

		mux := http.NewServeMux()

		// Add health endpoint
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if err := writeResponse(w, []byte("OK")); err != nil {
				// TODO: replace with logger
				logger.Get().Error("Failed to write response", "error", err)
			}
		})

		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sseServer.ServeHTTP(w, r)
		}))

		httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		}

		go func() {
			defer wg.Done()
			logger.Get().Info("Running KAgent Tools Server", "port", fmt.Sprintf(":%d", port))
			if err := httpServer.ListenAndServe(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					logger.Get().Error("Failed to start HTTP server", "error", err)
				} else {
					logger.Get().Info("HTTP server shut down gracefully")
				}
			}
		}()
	}
	go func() {
		<-signalChan
		logger.Get().Info("Received shutdown signal, shutting down...")
		cancel()

		if !stdio && httpServer != nil {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Get().Error("Failed to gracefully shutdown HTTP server", "error", err)
			}
		}
	}()

	wg.Wait()
	logger.Get().Info("Server shutdown complete.")
}

func runStdioServer(ctx context.Context, mcp *server.MCPServer) {
	logger.Get().Info("Running kyverno-tool-server STDIO:")
	stdioServer := server.NewStdioServer(mcp)
	if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		logger.Get().Error("Stdio server stopped", "error", err)
	}
}

func writeResponse(w http.ResponseWriter, data []byte) error {
	_, err := w.Write(data)
	return err
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
