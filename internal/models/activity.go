package models

import "time"

type Activity struct {
	BaseEntity
	UserID          int       `json:"userId" db:"user_id"`
	ActivityType    string    `json:"activityType" db:"activity_type"`
	Title           string    `json:"title" db:"title"`
	Description     string    `json:"description,omitempty" db:"description"`
	DurationMinutes int       `json:"durationMinutes,omitempty" db:"duration_minutes"`
	DistanceKm      float64   `json:"distanceKm,omitempty" db:"distance_km"`
	CaloriesBurned  int       `json:"caloriesBurned,omitempty" db:"calories_burned"`
	Notes           string    `json:"notes,omitempty" db:"notes"`
	ActivityDate    time.Time `json:"activityDate" db:"activity_date"`
}

type CreateActivityRequest struct {
	ActivityType    string    `json:"activityType"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"durationMinutes"`
	DistanceKm      float64   `json:"distanceKm"`
	CaloriesBurned  int       `json:"valoriesBurned"`
	Notes           string    `json:"notes"`
	ActivityDate    time.Time `json:"activityDate"`
}
