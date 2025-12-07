// Package main
package main

import (
	"context"
	"log"

	"github.com/benjamin-argo/lsa-poller/internal/config"
	"github.com/benjamin-argo/lsa-poller/internal/googleads"
	"github.com/benjamin-argo/lsa-poller/internal/gsheets"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded successfully")

	ctx := context.Background()

	// === GOOGLE SHEETS AUTH ===
	sheetsOAuthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{sheets.SpreadsheetsScope},
	}

	sheetsToken := &oauth2.Token{
		RefreshToken: cfg.SheetsRefreshToken, // ← Sheets token
	}

	sheetsHTTPClient := sheetsOAuthConfig.Client(ctx, sheetsToken)
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(sheetsHTTPClient))
	if err != nil {
		log.Fatalf("Failed to create sheets service: %v", err)
	}

	// Read spreadsheet
	resp, err := sheetsService.Spreadsheets.Values.Get(cfg.SheetID, "Sheet1!A:G").Do()
	if err != nil {
		log.Fatalf("Failed to read: %v", err)
	}

	log.Printf("Got %d rows", len(resp.Values))

	// Parse clients
	clients := gsheets.ParseClients(resp.Values)
	log.Printf("Found %d clients", len(clients))

	// === GOOGLE ADS CLIENT ===
	googleadsClient, err := googleads.NewClient(
		ctx,
		cfg.ClientID,
		cfg.ClientSecret,
		cfg.GoogleAdsRefreshToken, // ← Google Ads token
		cfg.GoogleAdsDeveloperToken,
		cfg.GoogleAdsManagerID,
	)
	if err != nil {
		log.Fatalf("Failed to create Google Ads client: %v", err)
	}
	if len(clients) > 0 {
		log.Printf("Attempting to fetch leads for account: %s (%s)", clients[0].AccountName, clients[0].AccountID)
	}

	// Fetch leads for first client
	if len(clients) > 0 {
		leads, err := googleadsClient.FetchLeads(ctx, clients[0].AccountID, 10)
		if err != nil {
			log.Fatalf("Failed to fetch leads: %v", err)
		}

		log.Printf("Got %d leads for %s", len(leads), clients[0].AccountName)
		for _, lead := range leads {
			log.Printf("  Lead ID: %s, Type: %s, Status: %s",
				lead.ID,
				lead.GetLeadTypeName(),
				lead.GetLeadStatusName())
		}
	}
}
