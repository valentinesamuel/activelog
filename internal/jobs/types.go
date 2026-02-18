package jobs

import "encoding/json"

// WelcomeEmailPayload is the data for sending a welcome email.
type WelcomeEmailPayload struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// WeeklySummaryPayload is the data for generating a weekly summary email.
type WeeklySummaryPayload struct {
	UserID int `json:"user_id"`
}

// ExportPayload is the data for generating a CSV/PDF export.
type ExportPayload struct {
	UserID int    `json:"user_id"`
	Format string `json:"format"` // "csv" or "pdf"
}

// mustMarshal marshals v to json.RawMessage or panics (used only in tests/examples).
func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
