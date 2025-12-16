// Package main
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nonstopautomation/lsa-poller/internal/config"
	"github.com/nonstopautomation/lsa-poller/internal/googleads"
	"github.com/nonstopautomation/lsa-poller/internal/gsheets"
	"github.com/nonstopautomation/lsa-poller/internal/webhook"
	"github.com/joho/godotenv"
)

func processLeads(client gsheets.Client, leads []googleads.Lead) {
	for _, lead := range leads {
		if err := webhook.SendLead(client.WebhookURL, lead); err != nil {
			log.Printf("Failed to send lead %s for client %s: %v",
				lead.ID, client.AccountName, err)
			continue
		}
		log.Printf("Sent lead %s for client %s", lead.ID, client.AccountName)
	}
}

func processClients(ctx context.Context, clients []gsheets.Client, adsClient *googleads.Client) {
	for _, client := range clients {
		// Add this check
		select {
		case <-ctx.Done():
			log.Println("Shutdown requested, stopping client processing")
			return
		default:
		}

		leads, err := adsClient.FetchLeads(ctx, client.AccountID, 100)

		if err != nil {
			log.Printf("[ERROR] Failed to fetch leads for %s (%s): %v",
				client.AccountName, client.AccountID, err)
		} else {
			if len(leads) > 0 {
				processLeads(client, leads)
			} else {
				log.Printf("[INFO] No new leads for %s", client.AccountName)
			}
		}

	}

	log.Printf("Finished processing leads this run for all clients.")
}

func startHealthServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		log.Printf("Starting health check server on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health check server error: %v", err)
		}
	}()

	return server
}

func getPort() int {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 8080
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid PORT value '%s', using default 8080", portStr)
		return 8080
	}
	return port
}

func run() error {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded successfully")

	ctx := context.Background()

	clients, err := gsheets.FetchClients(ctx, cfg)
	if err != nil {
		return fmt.Errorf("error fetching clients from spreadsheet: %w", err)
	}

	adsClient, err := googleads.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("error authenticating to googleads clients %w", err)
	}

	port := getPort()
	healthServer := startHealthServer(port)

	return runPoller(clients, adsClient, healthServer)
}

func runPoller(clients []gsheets.Client, adsClient *googleads.Client, healthServer *http.Server) error {
	pollerCtx, cancel := context.WithCancel(context.Background())

	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		cancel() // Cancels pollerCtx

		// Gracefully shutdown health server
		if healthServer != nil {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := healthServer.Shutdown(shutdownCtx); err != nil {
				log.Printf("Error shutting down health server: %v", err)
			} else {
				log.Println("Health server shut down gracefully")
			}
		}
	}()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("Starting initial poll...")
	processClients(pollerCtx, clients, adsClient) // Use pollerCtx

	for {
		select {
		case t := <-ticker.C:
			log.Printf("Proccessing started at %s", t.Format("15:04:05"))
			processClients(pollerCtx, clients, adsClient)

		case <-pollerCtx.Done():
			log.Println("Shutdown complete")
			return nil
		}
	}
}

func main() {
	fmt.Println("Starting Polling Process")
	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
