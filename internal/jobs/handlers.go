package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/valentinesamuel/activelog/internal/queue/types"
)

// HandleWelcomeEmail processes a welcome email job.
// Once an email provider is wired in, this will call emailProvider.Send.
func HandleWelcomeEmail(_ context.Context, payload types.JobPayload) error {
	var p WelcomeEmailPayload
	if err := json.Unmarshal(payload.Data, &p); err != nil {
		return fmt.Errorf("HandleWelcomeEmail: unmarshal: %w", err)
	}
	log.Printf("[job] welcome email -> userID=%d email=%s name=%s", p.UserID, p.Email, p.Name)
	return nil
}

// HandleWeeklySummary processes a weekly summary email job.
func HandleWeeklySummary(_ context.Context, payload types.JobPayload) error {
	var p WeeklySummaryPayload
	if err := json.Unmarshal(payload.Data, &p); err != nil {
		return fmt.Errorf("HandleWeeklySummary: unmarshal: %w", err)
	}
	log.Printf("[job] weekly summary -> userID=%d", p.UserID)
	return nil
}

// HandleGenerateExport processes a CSV/PDF export generation job.
func HandleGenerateExport(_ context.Context, payload types.JobPayload) error {
	var p ExportPayload
	if err := json.Unmarshal(payload.Data, &p); err != nil {
		return fmt.Errorf("HandleGenerateExport: unmarshal: %w", err)
	}
	log.Printf("[job] generate export -> userID=%d format=%s", p.UserID, p.Format)
	return nil
}
