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

type UserActivitySummary struct {
	Username       string `json:"username"`
	ActivityCount  int    `json:"activityCount"`
	UniqueTagCount int    `json:"uniqueTagCount"`
	ActivityLevel  string `json:"activityLevel,omitempty"` // Computed by service layer
	TotalActivities int   `json:"totalActivities"`         // Alias for ActivityCount
}

type TagUsage struct {
	TagName string `json:"tagName"`
	Count   int    `json:"count"`
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

func (sr *StatsRepository) GetWeeklyStats(ctx context.Context, userID int) (*WeeklyStats, error) {
	query := `
		SELECT
			COUNT(*)::int AS total_activities,
			COALESCE(SUM(duration_minutes), 0)::int AS total_duration,
			COALESCE(SUM(distance_km), 0)::float AS total_distance,
			COALESCE(AVG(duration_minutes), 0)::float AS avg_duration
		FROM activities
		WHERE user_id = $1
			AND activity_date >= NOW() - INTERVAL '7 days'
	`

	weeklyStats := &WeeklyStats{}

	row := sr.db.QueryRowContext(ctx, query, userID)

	err := row.Scan(
		&weeklyStats.TotalActivities,
		&weeklyStats.TotalDuration,
		&weeklyStats.TotalDistance,
		&weeklyStats.AvgDuration,
	)

	if err != nil {
		return nil, &errors.DatabaseError{
			Op:    "AGGREGATE",
			Table: "activities",
			Err:   err,
		}
	}

	return weeklyStats, nil
}

func (sr *StatsRepository) GetUserActivitySummary(ctx context.Context, userID int) (*UserActivitySummary, error) {
	query := `
		SELECT
			u.username,
			COUNT(DISTINCT a.id)::int        AS "activityCount",
			COUNT(DISTINCT t.id)::int        AS "uniqueTagCount"
		FROM users u
		LEFT JOIN activities a
			ON a.user_id = u.id
		LEFT JOIN activity_tags at
			ON at.activity_id = a.id
		LEFT JOIN tags t
			ON t.id = at.tag_id
		WHERE u.id = $1
		GROUP BY u.id, u.username;
	`

	row := sr.db.QueryRowContext(ctx, query, userID)

	userActivitySummary := &UserActivitySummary{}
	err := row.Scan(
		&userActivitySummary.Username,
		&userActivitySummary.ActivityCount,
		&userActivitySummary.UniqueTagCount,
	)

	if err != nil {
		return nil, &errors.DatabaseError{
			Op:    "AGGREGATE",
			Table: "activities",
			Err:   err,
		}
	}

	return userActivitySummary, nil
}

func (sr *StatsRepository) GetTopTagsByUser(ctx context.Context, userID int, limit int) ([]TagUsage, error) {
	query := `
		SELECT
			t.name AS tag_name,
			COUNT(*)::int AS usage_count
		FROM tags t
		INNER JOIN activity_tags at
			ON at.tag_id = t.id
		INNER JOIN activities a
			ON a.id = at.activity_id
		WHERE a.user_id = $1
		GROUP BY t.id, t.name
		ORDER BY usage_count DESC
		LIMIT $2
	`

	rows, err := sr.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, &errors.DatabaseError{
			Op:    "AGGREGATE",
			Table: "tags",
			Err:   err,
		}
	}
	defer rows.Close()

	var tagUsages []TagUsage
	for rows.Next() {
		var tagUsage TagUsage
		if err := rows.Scan(&tagUsage.TagName, &tagUsage.Count); err != nil {
			return nil, &errors.DatabaseError{
				Op:    "SCAN",
				Table: "tags",
				Err:   err,
			}
		}
		tagUsages = append(tagUsages, tagUsage)
	}

	if err := rows.Err(); err != nil {
		return nil, &errors.DatabaseError{
			Op:    "ITERATE",
			Table: "tags",
			Err:   err,
		}
	}

	// Return empty slice instead of nil if no tags found
	if tagUsages == nil {
		tagUsages = []TagUsage{}
	}

	return tagUsages, nil
}
