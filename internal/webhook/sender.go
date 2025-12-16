// Package webhook
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nonstopautomation/lsa-poller/internal/googleads"
)

func SendLead(webhookURL string, lead googleads.Lead) error {
	jsonBytes, err := json.Marshal(lead)
	if err != nil {
		return fmt.Errorf("failed to marshal lead: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second, // Don't wait forever
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Step 5: Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✓ Successfully sent lead %s to webhook\n", lead.ID)
	return nil
}
