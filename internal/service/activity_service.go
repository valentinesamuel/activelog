package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	appErrors "github.com/valentinesamuel/activelog/pkg/errors"
)

// ActivityService implements ActivityServiceInterface
// Encapsulates business logic for activity operations
type ActivityService struct {
	activityRepo repository.ActivityRepositoryInterface
	tagRepo      repository.TagRepositoryInterface
}

// NewActivityService creates a new activity service instance
func NewActivityService(
	activityRepo repository.ActivityRepositoryInterface,
	tagRepo repository.TagRepositoryInterface,
) *ActivityService {
	return &ActivityService{
		activityRepo: activityRepo,
		tagRepo:      tagRepo,
	}
}

// CreateActivity handles activity creation with business rules
func (s *ActivityService) CreateActivity(
	ctx context.Context,
	tx repository.TxConn,
	userID int,
	req *models.CreateActivityRequest,
) (*models.Activity, error) {
	// Business Rule 1: Activity date cannot be in the future
	if req.ActivityDate.After(time.Now()) {
		return nil, fmt.Errorf("activity date cannot be in the future")
	}

	// Business Rule 2: Duration must be reasonable (not more than 24 hours)
	if req.DurationMinutes > 1440 {
		return nil, fmt.Errorf("duration cannot exceed 24 hours (1440 minutes)")
	}

	// Business Rule 3: Distance must be positive if provided
	if req.DistanceKm < 0 {
		return nil, fmt.Errorf("distance must be positive")
	}

	// Build activity entity
	activity := &models.Activity{
		UserID:          userID,
		ActivityType:    req.ActivityType,
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		DistanceKm:      req.DistanceKm,
		CaloriesBurned:  req.CaloriesBurned,
		Notes:           req.Notes,
		ActivityDate:    req.ActivityDate,
	}

	// Create activity (tags support can be added later when needed)
	if err := s.activityRepo.Create(ctx, tx, activity); err != nil {
		log.Error().Err(err).Msg("Failed to create activity")
		return nil, err
	}

	log.Info().
		Int("user_id", userID).
		Int64("activity_id", activity.ID).
		Str("type", activity.ActivityType).
		Msg("Activity created successfully")

	return activity, nil
}

// UpdateActivity handles activity updates with business rules
func (s *ActivityService) UpdateActivity(
	ctx context.Context,
	tx repository.TxConn,
	userID int,
	activityID int,
	req *models.UpdateActivityRequest,
) (*models.Activity, error) {
	// Business Rule 1: Verify activity exists and belongs to user
	existingActivity, err := s.activityRepo.GetByID(ctx, int64(activityID))
	if err != nil {
		return nil, appErrors.ErrNotFound
	}

	// Business Rule 2: Verify ownership
	if existingActivity.UserID != userID {
		return nil, appErrors.ErrUnauthorized
	}

	// Business Rule 3: Activity date cannot be in the future
	if req.ActivityDate != nil && req.ActivityDate.After(time.Now()) {
		return nil, fmt.Errorf("activity date cannot be in the future")
	}

	// Business Rule 4: Duration must be reasonable
	if req.DurationMinutes != nil && *req.DurationMinutes > 1440 {
		return nil, fmt.Errorf("duration cannot exceed 24 hours (1440 minutes)")
	}

	// Business Rule 5: Distance must be positive if provided
	if req.DistanceKm != nil && *req.DistanceKm < 0 {
		return nil, fmt.Errorf("distance must be positive")
	}

	// Apply partial updates to existing activity
	if req.ActivityType != nil {
		existingActivity.ActivityType = *req.ActivityType
	}
	if req.Title != nil {
		existingActivity.Title = *req.Title
	}
	if req.Description != nil {
		existingActivity.Description = *req.Description
	}
	if req.DurationMinutes != nil {
		existingActivity.DurationMinutes = *req.DurationMinutes
	}
	if req.DistanceKm != nil {
		existingActivity.DistanceKm = *req.DistanceKm
	}
	if req.CaloriesBurned != nil {
		existingActivity.CaloriesBurned = *req.CaloriesBurned
	}
	if req.Notes != nil {
		existingActivity.Notes = *req.Notes
	}
	if req.ActivityDate != nil {
		existingActivity.ActivityDate = *req.ActivityDate
	}

	// Perform update
	if err := s.activityRepo.Update(ctx, tx, activityID, existingActivity); err != nil {
		log.Error().Err(err).Int("activity_id", activityID).Msg("Failed to update activity")
		return nil, err
	}

	// Fetch updated activity to return
	updated, err := s.activityRepo.GetByID(ctx, int64(activityID))
	if err != nil {
		return nil, err
	}

	log.Info().
		Int("user_id", userID).
		Int("activity_id", activityID).
		Msg("Activity updated successfully")

	return updated, nil
}

// DeleteActivity handles activity deletion with business rules
func (s *ActivityService) DeleteActivity(
	ctx context.Context,
	tx repository.TxConn,
	userID int,
	activityID int,
) error {
	// Business Rule 1: Verify activity exists and belongs to user
	existingActivity, err := s.activityRepo.GetByID(ctx, int64(activityID))
	if err != nil {
		return appErrors.ErrNotFound
	}

	// Business Rule 2: Verify ownership
	if existingActivity.UserID != userID {
		return appErrors.ErrUnauthorized
	}

	// Business Rule 3: Prevent deletion of activities older than 1 year (business policy)
	// This is an example business rule - you may want to remove or modify this
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	if existingActivity.CreatedAt.Before(oneYearAgo) {
		log.Warn().
			Int("activity_id", activityID).
			Time("created_at", existingActivity.CreatedAt).
			Msg("Attempted to delete activity older than 1 year")
		// For now, we'll allow it but log a warning
		// You could return an error here to enforce the rule
	}

	// Perform deletion (repository handles cascade)
	if err := s.activityRepo.Delete(ctx, tx, activityID, userID); err != nil {
		log.Error().Err(err).Int("activity_id", activityID).Msg("Failed to delete activity")
		return err
	}

	log.Info().
		Int("user_id", userID).
		Int("activity_id", activityID).
		Msg("Activity deleted successfully")

	return nil
}
