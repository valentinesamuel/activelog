package handlers

import (
	"net/http"

	"github.com/valentinesamuel/activelog/internal/featureflags"
	"github.com/valentinesamuel/activelog/pkg/response"
)

// FeaturesHandler serves the current feature flag state
type FeaturesHandler struct {
	flags *featureflags.FeatureFlags
}

// NewFeaturesHandler creates a FeaturesHandler
func NewFeaturesHandler(flags *featureflags.FeatureFlags) *FeaturesHandler {
	return &FeaturesHandler{flags: flags}
}

// GetFeatures returns the feature flag state for the requesting user
func (h *FeaturesHandler) GetFeatures(w http.ResponseWriter, r *http.Request) {
	features := map[string]bool{
		"comments": h.flags.IsEnabled("comments"),
		"likes":    h.flags.IsEnabled("likes"),
		"friends":  h.flags.IsEnabled("friends"),
		"webhooks": h.flags.IsEnabled("webhooks"),
		"feed":     h.flags.IsEnabled("feed"),
	}
	response.SendJSON(w, http.StatusOK, features)
}
