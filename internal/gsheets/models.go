// Package sheets
package gsheets

type Client struct {
	RowIndex       int
	AccountID      string
	AccountName    string
	WebhookURL     string
	Status         string
	LastPoll       string
	LeadsProcessed string
	LastError      string
}

// ParseClients converts spreadsheet rows into Client structs
func ParseClients(rows [][]any) []Client {
	if len(rows) <= 1 {
		// Only headers or empty
		return []Client{}
	}

	clients := []Client{} // ← This is a slice!

	// Skip header row (index 0), start at row 1
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		client := Client{
			RowIndex: i + 1, // 1-based for spreadsheet
		}

		// Safely get each column
		if len(row) > 0 {
			client.AccountID = getString(row[0])
		}
		if len(row) > 1 {
			client.AccountName = getString(row[1])
		}
		if len(row) > 2 {
			client.WebhookURL = getString(row[2])
		}
		if len(row) > 3 {
			client.Status = getString(row[3])
		}
		if len(row) > 4 {
			client.LastPoll = getString(row[4])
		}
		if len(row) > 5 {
			client.LeadsProcessed = getString(row[5])
		}
		if len(row) > 6 {
			client.LastError = getString(row[6])
		}

		clients = append(clients, client) // ← Add to slice
	}

	return clients
}

// Helper to safely convert interface{} to string
func getString(val any) string {
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func SafeGetString(value any) string {
	if value == nil {
		return ""
	}

	if str, ok := value.(string); ok {
		return str
	}

	return ""
}
