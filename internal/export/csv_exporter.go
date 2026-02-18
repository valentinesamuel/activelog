package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/valentinesamuel/activelog/internal/models"
)

// ExportActivitiesCSV streams activities as CSV to w.
// It writes a header row followed by one row per activity.
func ExportActivitiesCSV(_ context.Context, activities []*models.Activity, w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header row
	header := []string{
		"id", "user_id", "activity_type", "title", "description",
		"duration_minutes", "distance_km", "calories_burned",
		"notes", "activity_date", "created_at",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write each activity as a row
	for _, a := range activities {
		row := []string{
			fmt.Sprintf("%d", a.ID),
			fmt.Sprintf("%d", a.UserID),
			a.ActivityType,
			a.Title,
			a.Description,
			fmt.Sprintf("%d", a.DurationMinutes),
			fmt.Sprintf("%.2f", a.DistanceKm),
			fmt.Sprintf("%d", a.CaloriesBurned),
			a.Notes,
			a.ActivityDate.Format("2006-01-02"),
			a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
