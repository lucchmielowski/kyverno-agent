package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"kyverno-agent/pkg/tools"
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

	// Ctx setup for raceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
				fmt.Println("Failed to write response: ", err)
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
			fmt.Printf("Starting HTTP server on port %d\n", port)
			if err := httpServer.ListenAndServe(); err != nil {
				if !errors.Is(err, http.ErrServerClosed) {
					fmt.Println("Failed to start HTTP server: ", err)
				} else {
					fmt.Println("HTTP server stopped gracefully")
				}
			}
		}()
	}
	go func() {
		<-signalChan
		fmt.Println("Received signal, shutting down...")
		cancel()

		if !stdio && httpServer != nil {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()

			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				fmt.Println("Failed to shutdown HTTP server: ", err)
			}
		}
	}()

	wg.Wait()
	fmt.Println("Server shutdown complete")
}

func runStdioServer(ctx context.Context, mcp *server.MCPServer) {
	fmt.Println("Starting stdio server")
	stdioServer := server.NewStdioServer(mcp)
	if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		fmt.Println("Failed to start stdio server: ", err)
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
