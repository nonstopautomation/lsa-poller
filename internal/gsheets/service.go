package gsheets

import (
	"context"
	"fmt"

	"github.com/nonstopautomation/lsa-poller/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// NewService creates an authenticated Google Sheets service
func NewService(ctx context.Context, cfg *config.ProjectConfig) (*sheets.Service, error) {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{sheets.SpreadsheetsScope},
	}

	token := &oauth2.Token{
		RefreshToken: cfg.SheetsRefreshToken,
	}

	httpClient := oauthConfig.Client(ctx, token)

	return sheets.NewService(ctx, option.WithHTTPClient(httpClient))
}

// FetchClients reads and parses client list from spreadsheet
func FetchClients(ctx context.Context, cfg *config.ProjectConfig) ([]Client, error) {
	service, err := NewService(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	resp, err := service.Spreadsheets.Values.Get(cfg.SheetID, "Sheet1!A:G").Do()
	if err != nil {
		return nil, fmt.Errorf("read spreadsheet: %w", err)
	}

	return ParseClients(resp.Values), nil
}
