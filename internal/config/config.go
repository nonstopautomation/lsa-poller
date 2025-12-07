// Package config
package config

import (
	"fmt"
	"os"
)

type ProjectConfig struct {
	RefreshToken            string
	SheetID                 string
	SheetName               string
	ClientSecret            string
	ClientID                string
	GoogleAdsManagerID      string
	GoogleAdsDeveloperToken string
}

func Load() (*ProjectConfig, error) {
	cfg := &ProjectConfig{
		RefreshToken:            os.Getenv("GOOGLE_SHEETS_REFRESH_TOKEN"),
		ClientSecret:            os.Getenv("GOOGLE_CLIENT_SECRET"),
		ClientID:                os.Getenv("GOOGLE_CLIENT_ID"),
		SheetID:                 os.Getenv("SHEET_ID"),
		SheetName:               os.Getenv("SHEET_NAME"),
		GoogleAdsManagerID:      os.Getenv("GOOGLE_ADS_MANAGER_ID"),
		GoogleAdsDeveloperToken: os.Getenv("GOOGLE_ADS_DEVELOPER_TOKEN"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *ProjectConfig) Validate() error {
	checks := map[string]string{
		c.RefreshToken: "GOOGLE_SHEETS_REFRESH_TOKEN",
		c.SheetID:      "SHEET_ID",
		c.SheetName:    "SHEET_NAME",
		c.ClientSecret: "GOOGLE_CLIENT_SECRET",
		c.ClientID:     "GOOGLE_CLIENT_ID",
	}

	for value, name := range checks {
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}
