package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/validator"
	"github.com/valentinesamuel/activelog/pkg/auth"
	"github.com/valentinesamuel/activelog/pkg/response"
)

type UserHandler struct {
	repo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		repo: repo,
	}
}

func (ua *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var requestPayload models.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := validator.Validate(&requestPayload)
	if err != nil {
		validationError := validator.FormatValidationErrors(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  "validation failed",
			"fields": validationError,
		})
		return
	}

	user := &models.User{
		Email:    requestPayload.Email,
		Username: requestPayload.Username,
	}

	encodedHash, err := auth.HashPassword(requestPayload.Password)
	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to hash password")
		response.Error(w, http.StatusInternalServerError, "Invalid password")
		return
	}

	user.PasswordHash = encodedHash

	if err := ua.repo.CreateUser(ctx, user); err != nil {
		log.Error().Err(err).Msg("❌ Failed to create user")
		response.Error(w, http.StatusInternalServerError, "❌ Failed to create user")
		return
	}

	log.Info().Str("email", user.Email).Msg("✅ Activity Created")
	response.SendJSON(w, http.StatusOK, map[string]string{
		"email":    user.Email,
		"username": user.Username,
	})
}
