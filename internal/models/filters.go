package models

import "time"

type ActivityFilters struct {
	ActivityType string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}
