package repository

import (
	"context"
	"encoding/json"
	"github.com/valentinesamuel/activelog/pkg/errors"
)

type StatsRepository struct {
	db DBConn
}

type MonthlyStats map[string]int

type WeeklyStats struct {
	TotalActivities int     `json:"totalActivities"`
	TotalDuration   int     `json:"totalDurationMinutes"`
	TotalDistance   float64 `json:"totalDistanceKm"`
	AvgDuration     float64 `json:"avgDurationMinutes"`
}

func NewStatsRepository(db DBConn) *StatsRepository {
	return &StatsRepository{
		db: db,
	}
}

func (sr *StatsRepository) GetMonthlyStats(ctx context.Context, userID int) (*MonthlyStats, error) {
	query := `
		SELECT COALESCE(
			json_object_agg(activity_type, activity_count),
			'{}'::json
		) as stats
		FROM (
			SELECT
				activity_type,
				COUNT(*)::int as activity_count
			FROM activities
			WHERE user_id = $1
				AND activity_date >= NOW() - INTERVAL '30 days'
			GROUP BY activity_type
		) as activity_stats
	`

	monthlyStats := &MonthlyStats{}

	row := sr.db.QueryRowContext(ctx, query, userID)

	var statsJSON []byte
	if err := row.Scan(&statsJSON); err != nil {
		return nil, &errors.DatabaseError{
			Op:    "AGGREGATE",
			Table: "activities",
			Err:   err,
		}
	}

	// Unmarshal JSON into map
	if len(statsJSON) > 0 {
		if err := json.Unmarshal(statsJSON, monthlyStats); err != nil {
			return nil, &errors.DatabaseError{
				Op:    "AGGREGATE",
				Table: "activities",
				Err:   err,
			}
		}
	}

	return monthlyStats, nil
}

func (sr *StatsRepository) GetActivityCountByType(ctx context.Context, userID int) (map[string]int, error) {
	query := `
		SELECT COALESCE(
			json_object_agg(activity_type, activity_count),
			'{}'::json
		) as stats
		FROM (
			SELECT
				activity_type,
				COUNT(*)::int as activity_count
			FROM activities
			WHERE user_id = $1
			GROUP BY activity_type
		) as activity_stats
	`

	row := sr.db.QueryRowContext(ctx, query, userID)

	var statsJSON []byte
	if err := row.Scan(&statsJSON); err != nil {
		return nil, &errors.DatabaseError{
			Op:    "AGGREGATE",
			Table: "activities",
			Err:   err,
		}
	}

	stats := make(map[string]int)

	if err := json.Unmarshal(statsJSON, &stats); err != nil {
		return nil, &errors.DatabaseError{
			Op:    "AGGREGATE",
			Table: "activities",
			Err:   err,
		}
	}

	return stats, nil
}
