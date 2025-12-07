// Package googleads
package googleads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Client struct {
	httpClient       *http.Client
	developerToken   string
	managerAccountID string
}

type Lead struct {
	ID               string
	ContactDetails   ContactDetails
	LeadType         int
	LeadStatus       int
	CategoryID       string
	ServiceID        string
	CreationDateTime string
	LeadCharged      bool
}

// ContactDetails contains the lead's contact information
type ContactDetails struct {
	PhoneNumber  string
	Email        string
	ConsumerName string
}

var LeadTypes = map[int]string{
	1: "Booking",
	2: "Message",
	3: "Phone Call",
	4: "Unknown",
}

// LeadStatuses mappings
var LeadStatuses = map[int]string{
	1:  "Active",
	2:  "Booked",
	3:  "Consumer Declined",
	4:  "Declined",
	5:  "Disabled",
	6:  "Expired",
	7:  "New",
	8:  "Unknown",
	9:  "Unspecified",
	10: "Wiped Out",
}

// GetLeadTypeName returns the human-readable lead type
func (l *Lead) GetLeadTypeName() string {
	if name, ok := LeadTypes[l.LeadType]; ok {
		return name
	}
	return "Unknown"
}

// GetLeadStatusName returns the human-readable lead status
func (l *Lead) GetLeadStatusName() string {
	if name, ok := LeadStatuses[l.LeadStatus]; ok {
		return name
	}
	return "Unknown"
}

// NewClient creates a new Google Ads API client
func NewClient(ctx context.Context, clientID, clientSecret, refreshToken, developerToken, managerAccountID string) (*Client, error) {
	// Set up OAuth2 (same as Sheets!)
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/adwords"},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	httpClient := config.Client(ctx, token)

	return &Client{
		httpClient:       httpClient,
		developerToken:   developerToken,
		managerAccountID: managerAccountID,
	}, nil
}

// FetchLeads gets recent leads for an account
func (c *Client) FetchLeads(ctx context.Context, accountID string, maxLeads int) ([]Lead, error) {
	// Build the API URL

	url := fmt.Sprintf("https://googleads.googleapis.com/v22/customers/%s/googleAds:searchStream", accountID)
	print("FETCHING LEADS")

	// Build the GAQL query
	query := fmt.Sprintf(`
		SELECT 
			local_services_lead.id,
			local_services_lead.contact_details,
			local_services_lead.lead_type,
			local_services_lead.lead_status,
			local_services_lead.category_id,
			local_services_lead.service_id,
			local_services_lead.creation_date_time,
			local_services_lead.lead_charged
		FROM local_services_lead
		ORDER BY local_services_lead.creation_date_time DESC
		LIMIT %d
	`, maxLeads)

	// Create the request body
	requestBody := map[string]string{
		"query": query,
	}

	// Convert to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("developer-token", c.developerToken)
	req.Header.Set("login-customer-id", c.managerAccountID)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// TODO: Parse the response into Lead structs
	// For now, just print it
	fmt.Printf("Response: %s\n", string(body))

	return []Lead{}, nil
}
