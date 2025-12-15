// Package googleads
package googleads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/benjamin-argo/lsa-poller/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Client struct {
	httpClient       *http.Client
	developerToken   string
	managerAccountID string
}

type SearchResults struct {
	Results []localServicesLead `json:"results"`
}

type localServicesLead struct {
	LocalServicesLead Lead `json:"localServicesLead"`
}

type Lead struct {
	ID               string         `json:"id"`
	ResourceName     string         `json:"resourceName"`
	ContactDetails   ContactDetails `json:"contactDetails"`
	LeadType         string         `json:"leadType"`
	LeadStatus       string         `json:"leadStatus"`
	CategoryID       string         `json:"categoryId"`
	ServiceID        string         `json:"serviceId"`
	CreationDateTime string         `json:"creationDateTime"`
	LeadCharged      bool           `json:"leadCharged"`
	AccountTimezone  string         `json:"accountTimezone"`
}

type ContactDetails struct {
	PhoneNumber  string `json:"phoneNumber"`
	Email        string `json:"email"`
	ConsumerName string `json:"consumerName"`
}

func NewClient(ctx context.Context, cfg *config.ProjectConfig) (*Client, error) {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/adwords"},
	}

	token := &oauth2.Token{
		RefreshToken: cfg.GoogleAdsRefreshToken,
	}

	// Use oauthConfig.Client(), not config.Client()
	httpClient := oauthConfig.Client(ctx, token)

	return &Client{
		httpClient:       httpClient,
		developerToken:   cfg.GoogleAdsDeveloperToken,
		managerAccountID: cfg.GoogleAdsManagerID,
	}, nil
}

func (c *Client) GetAccountTimezone(ctx context.Context, accountID string) (string, error) {
	url := fmt.Sprintf("https://googleads.googleapis.com/v22/customers/%s/googleAds:searchStream", accountID)

	query := `
        SELECT customer.time_zone
        FROM customer
        LIMIT 1
    `

	requestBody := map[string]string{
		"query": query,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("developer-token", c.developerToken)
	req.Header.Set("login-customer-id", c.managerAccountID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the response to extract timezone
	var response []struct {
		Results []struct {
			Customer struct {
				TimeZone string `json:"timeZone"`
			} `json:"customer"`
		} `json:"results"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response) > 0 && len(response[0].Results) > 0 {
		return response[0].Results[0].Customer.TimeZone, nil
	}

	return "", fmt.Errorf("no timezone found in response")
}

// FetchLeads gets recent leads for an account
func (c *Client) FetchLeads(ctx context.Context, accountID string, maxLeads int) ([]Lead, error) {
	url := fmt.Sprintf("https://googleads.googleapis.com/v22/customers/%s/googleAds:searchStream", accountID)

	timezone, err := c.GetAccountTimezone(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get timezone: %w", err)
	}

	// Get leads from last hour in account's timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	cutoffTime := time.Now().In(loc).Add(-1 * time.Hour)
	cutoffStr := cutoffTime.Format("2006-01-02 15:04:05")

	fmt.Printf("Fetching leads created after %s (%s) for %s\n", cutoffStr, timezone, accountID)

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
        WHERE local_services_lead.creation_date_time >= '%s'
        ORDER BY local_services_lead.creation_date_time DESC
        LIMIT %d
    `, cutoffStr, maxLeads)

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

	leads, err := ParseLeads(body)
	if err != nil {
		return nil, fmt.Errorf("couldnt parse leads %w", err)
	}

	return leads, nil
}

func ParseLeads(body []byte) ([]Lead, error) {
	var response []SearchResults

	err := json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("couldnt unmarshall input: %w", err)
	}

	var leads []Lead

	for _, resp := range response {
		for _, lead := range resp.Results {
			if lead.LocalServicesLead.ContactDetails.Email != "" || lead.LocalServicesLead.ContactDetails.PhoneNumber != "" {
				leads = append(leads, lead.LocalServicesLead)
			} else {
				fmt.Println("FOUND A LEAD WITHOUT AN EMAIL OR NUMBER")
				prettyJSON, _ := json.MarshalIndent(lead.LocalServicesLead, "", "  ")
				fmt.Println(string(prettyJSON))
			}
		}
	}

	return leads, nil
}
