package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/validator"
	"github.com/valentinesamuel/activelog/pkg/auth"
	appErrors "github.com/valentinesamuel/activelog/pkg/errors"
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

	existingUser, error := ua.repo.FindUserByEmail(ctx, requestPayload.Email)

	if existingUser != nil {
		log.Error().Err(error).Str("email", requestPayload.Email).Msg("User already exists")
		response.Error(w, http.StatusBadRequest, "User already exists")
		return
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
	response.SendJSON(w, http.StatusCreated, map[string]map[string]string{
		"user": {
			"email":    user.Email,
			"username": user.Username,
		},
	})
}

func (ua *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var requestPayload models.LoginUserRequest

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

	user, err := ua.repo.FindUserByEmail(ctx, requestPayload.Email)

	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}

		log.Error().Err(err).Str("email", requestPayload.Email).Msg("User not found")
		response.Error(w, http.StatusInternalServerError, "Invalid Credentials")
		return
	}

	passwordMatch, err := auth.VerifyPassword(requestPayload.Password, user.PasswordHash)

	if err != nil {
		log.Error().Err(err).Msg("❌ Password comparison failed")
		response.Error(w, http.StatusInternalServerError, "Invalid Credentials")
		return
	}

	if !passwordMatch {
		log.Error().Err(err).Msg("❌ Password mismatch")
		response.Error(w, http.StatusInternalServerError, "Invalid credentials")
		return
	}

	token, err := auth.GenerateJwtToken(int(user.ID), user.Email)
	if err != nil {
		log.Error().Err(err).Msg("❌ Failed to generate jwt")
		response.Error(w, http.StatusInternalServerError, "Server error")
		return
	}

	response.SendJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"email": user.Email,
	})
}
