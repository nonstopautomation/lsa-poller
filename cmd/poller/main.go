// Package main
package main

import (
	"context"
	"fmt"
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

	ctx := context.Background()

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{sheets.SpreadsheetsScope},
	}

	token := &oauth2.Token{
		RefreshToken: cfg.RefreshToken,
	}

	httpClient := oauthConfig.Client(ctx, token)

	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatalf("Failed to create sheets service: %v", err)
	}

	// Read the spreadsheet
	resp, err := sheetsService.Spreadsheets.Values.Get(cfg.SheetID, "Sheet1!A:G").Do()
	if err != nil {
		log.Fatalf("Failed to read: %v", err)
	}

	clients := gsheets.ParseClients(resp.Values)

	googleadsClient, err := googleads.NewClient(ctx, cfg.ClientID, cfg.ClientSecret, cfg.RefreshToken, cfg.GoogleAdsDeveloperToken, cfg.GoogleAdsManagerID)
	if err != nil {
		leads, err := googleadsClient.FetchLeads(ctx, clients[0].AccountID, 10)
		fmt.Println(err)
		fmt.Println(leads)
	}
}
