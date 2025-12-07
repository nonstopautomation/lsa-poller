// Package config handles application configuration from environment variables
package config

import (
	"fmt"
	"os"
)

type ProjectConfig struct {
	// Shared OAuth credentials
	ClientID     string
	ClientSecret string

	// Google Sheets
	SheetsRefreshToken string
	SheetID            string
	SheetName          string

	// Google Ads
	GoogleAdsRefreshToken   string
	GoogleAdsDeveloperToken string
	GoogleAdsManagerID      string
}

func Load() (*ProjectConfig, error) {
	cfg := &ProjectConfig{
		// Shared
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),

		// Sheets
		SheetsRefreshToken: os.Getenv("GOOGLE_SHEETS_REFRESH_TOKEN"),
		SheetID:            os.Getenv("SHEET_ID"),
		SheetName:          os.Getenv("SHEET_NAME"),

		// Google Ads
		GoogleAdsRefreshToken:   os.Getenv("GOOGLE_ADS_REFRESH_TOKEN"),
		GoogleAdsDeveloperToken: os.Getenv("GOOGLE_ADS_DEVELOPER_TOKEN"),
		GoogleAdsManagerID:      os.Getenv("GOOGLE_ADS_MANAGER_ACCOUNT_ID"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *ProjectConfig) Validate() error {
	checks := map[string]string{
		c.ClientID:                "GOOGLE_CLIENT_ID",
		c.ClientSecret:            "GOOGLE_CLIENT_SECRET",
		c.SheetsRefreshToken:      "GOOGLE_SHEETS_REFRESH_TOKEN",
		c.SheetID:                 "SHEET_ID",
		c.SheetName:               "SHEET_NAME",
		c.GoogleAdsRefreshToken:   "GOOGLE_ADS_REFRESH_TOKEN",
		c.GoogleAdsDeveloperToken: "GOOGLE_ADS_DEVELOPER_TOKEN",
		c.GoogleAdsManagerID:      "GOOGLE_ADS_MANAGER_ACCOUNT_ID",
	}

	for value, name := range checks {
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}

	return nil
}
