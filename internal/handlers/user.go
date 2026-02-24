package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/valentinesamuel/activelog/internal/models"
	"github.com/valentinesamuel/activelog/internal/repository"
	"github.com/valentinesamuel/activelog/internal/platform/validator"
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
		response.Fail(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := validator.Validate(&requestPayload)
	if err != nil {
		response.ValidationFail(w, r, validator.FormatValidationErrors(err))
		return
	}

	user := &models.User{
		Email:    requestPayload.Email,
		Username: requestPayload.Username,
	}

	existingUser, err := ua.repo.FindUserByEmail(ctx, requestPayload.Email)

	if existingUser != nil {
		log.Error().Err(err).Str("email", requestPayload.Email).Msg("User already exists")
		response.Fail(w, r, http.StatusBadRequest, "User already exists")
		return
	}

	encodedHash, err := auth.HashPassword(requestPayload.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		response.Fail(w, r, http.StatusInternalServerError, "Invalid password")
		return
	}

	user.PasswordHash = encodedHash

	if err := ua.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, appErrors.ErrAlreadyExists) {
			response.Fail(w, r, http.StatusConflict, "User already exists")
			return
		}
		log.Error().Err(err).Msg("Failed to create user")
		response.Fail(w, r, http.StatusInternalServerError, "Failed to create user")
		return
	}

	log.Info().Str("email", user.Email).Msg("Activity Created")
	response.Success(w, r, http.StatusCreated, map[string]map[string]string{
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
		response.Fail(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := validator.Validate(&requestPayload)
	if err != nil {
		response.ValidationFail(w, r, validator.FormatValidationErrors(err))
		return
	}

	user, err := ua.repo.FindUserByEmail(ctx, requestPayload.Email)

	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			response.Fail(w, r, http.StatusNotFound, "User not found")
			return
		}

		log.Error().Err(err).Str("email", requestPayload.Email).Msg("User not found")
		response.Fail(w, r, http.StatusInternalServerError, "Invalid Credentials")
		return
	}

	passwordMatch, err := auth.VerifyPassword(requestPayload.Password, user.PasswordHash)

	if err != nil {
		log.Error().Err(err).Msg("Password comparison failed")
		response.Fail(w, r, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !passwordMatch {
		log.Warn().Msg("Password mismatch")
		response.Fail(w, r, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := auth.GenerateJwtToken(int(user.ID), user.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate jwt")
		response.Fail(w, r, http.StatusInternalServerError, "Server error")
		return
	}

	response.Success(w, r, http.StatusOK, map[string]interface{}{
		"token": token,
		"email": user.Email,
	})
}
