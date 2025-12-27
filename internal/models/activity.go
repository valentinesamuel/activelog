package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

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
	ActivityType    string    `json:"activityType" validate:"required,min=2,max=50"`
	Title           string    `json:"title" validate:"max=255"`
	Description     string    `json:"description" validate:"max=1000"`
	DurationMinutes int       `json:"durationMinutes" validate:"omitempty,min=1,max=1440"`
	DistanceKm      float64   `json:"distanceKm" validate:"omitempty,min=0"`
	CaloriesBurned  int       `json:"caloriesBurned" validate:"omitempty,min=0"`
	Notes           string    `json:"notes" validate:"max=2000"`
	ActivityDate    time.Time `json:"activityDate" validate:"required"`
}

func (r *CreateActivityRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
