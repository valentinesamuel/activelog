package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type Activity struct {
	BaseEntity
	UserID          int       `json:"userId" `
	ActivityType    string    `json:"activityType" `
	Title           string    `json:"title" `
	Description     string    `json:"description,omitempty" `
	DurationMinutes int       `json:"durationMinutes,omitempty" `
	DistanceKm      float64   `json:"distanceKm,omitempty" `
	CaloriesBurned  int       `json:"caloriesBurned,omitempty" `
	Notes           string    `json:"notes,omitempty" `
	ActivityDate    time.Time `json:"activityDate" `
	Tags            []string  `json:"tags,omitempty" `
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

type UpdateActivityRequest struct {
	ActivityType    *string    `json:"activityType" validate:"omitempty,min=2,max=50"`
	Title           *string    `json:"title" validate:"omitempty,max=255"`
	Description     *string    `json:"description" validate:"omitempty,max=1000"`
	DurationMinutes *int       `json:"durationMinutes" validate:"omitempty,min=1,max=1440"`
	DistanceKm      *float64   `json:"distanceKm" validate:"omitempty,min=0"`
	CaloriesBurned  *int       `json:"caloriesBurned" validate:"omitempty,min=0"`
	Notes           *string    `json:"notes" validate:"omitempty,max=2000"`
	ActivityDate    *time.Time `json:"activityDate"`
}

func (r *CreateActivityRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
