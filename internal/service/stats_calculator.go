package service

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// StatsCalculator aggregates daily activity stats into the daily_stats table.
type StatsCalculator struct {
	db *sql.DB
}

// NewStatsCalculator creates a StatsCalculator backed by a raw *sql.DB.
func NewStatsCalculator(db *sql.DB) *StatsCalculator {
	return &StatsCalculator{db: db}
}

// CalculateDailyStats aggregates the previous day's activities for every user
// and upserts the results into the daily_stats table.
func (s *StatsCalculator) CalculateDailyStats(ctx context.Context) error {
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

	query := `
		INSERT INTO daily_stats (user_id, date, total_activities, total_distance_km, total_duration_minutes)
		SELECT
			user_id,
			$1::date                                AS date,
			COUNT(*)::int                           AS total_activities,
			COALESCE(SUM(distance_km), 0)::numeric  AS total_distance_km,
			COALESCE(SUM(duration_minutes), 0)::int AS total_duration_minutes
		FROM activities
		WHERE DATE(activity_date AT TIME ZONE 'UTC') = $1::date
		  AND deleted_at IS NULL
		GROUP BY user_id
		ON CONFLICT (user_id, date) DO UPDATE SET
			total_activities       = EXCLUDED.total_activities,
			total_distance_km      = EXCLUDED.total_distance_km,
			total_duration_minutes = EXCLUDED.total_duration_minutes
	`

	result, err := s.db.ExecContext(ctx, query, yesterday)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	log.Printf("[scheduler] daily_stats: upserted %d rows for %s", rows, yesterday)
	return nil
}
