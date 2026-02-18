package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/go-pdf/fpdf"
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

// GenerateActivityReport generates a PDF report for the given activities.
// It includes a summary section and a table of all activities.
func GenerateActivityReport(_ context.Context, activities []*models.Activity) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 12, "Activity Report", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Summary section
	totalCount := len(activities)
	var totalDuration int
	var totalDistance float64
	for _, a := range activities {
		totalDuration += a.DurationMinutes
		totalDistance += a.DistanceKm
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Summary")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, fmt.Sprintf("Total Activities: %d", totalCount))
	pdf.Ln(7)
	pdf.Cell(0, 7, fmt.Sprintf("Total Duration: %d minutes", totalDuration))
	pdf.Ln(7)
	pdf.Cell(0, 7, fmt.Sprintf("Total Distance: %.2f km", totalDistance))
	pdf.Ln(12)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	colWidths := []float64{25, 30, 50, 30, 30}
	headers := []string{"Date", "Type", "Title", "Duration (min)", "Distance (km)"}
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 9)
	for _, a := range activities {
		pdf.CellFormat(colWidths[0], 7, a.ActivityDate.Format("2006-01-02"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[1], 7, truncateString(a.ActivityType, 15), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[2], 7, truncateString(a.Title, 28), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[3], 7, fmt.Sprintf("%d", a.DurationMinutes), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[4], 7, fmt.Sprintf("%.2f", a.DistanceKm), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	// Output to bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
