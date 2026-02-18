package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	requestcontext "github.com/valentinesamuel/activelog/internal/requestContext"
	"github.com/valentinesamuel/activelog/internal/repository"
	webhookTypes "github.com/valentinesamuel/activelog/internal/webhook/types"
	"github.com/valentinesamuel/activelog/pkg/response"
)

// WebhookHandler handles webhook registration endpoints
type WebhookHandler struct {
	webhookRepo *repository.WebhookRepository
}

// NewWebhookHandler creates a new WebhookHandler
func NewWebhookHandler(webhookRepo *repository.WebhookRepository) *WebhookHandler {
	return &WebhookHandler{webhookRepo: webhookRepo}
}

type createWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// CreateWebhook handles POST /api/v1/webhooks
func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := requestcontext.FromContext(ctx)

	var req createWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.URL == "" {
		response.Error(w, http.StatusBadRequest, "URL is required")
		return
	}
	if len(req.Events) == 0 {
		response.Error(w, http.StatusBadRequest, "At least one event is required")
		return
	}

	secret, err := generateSecret()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate webhook secret")
		return
	}

	wh := &webhookTypes.Webhook{
		UserID: user.Id,
		URL:    req.URL,
		Events: req.Events,
		Secret: secret,
		Active: true,
	}
	if err := h.webhookRepo.Create(ctx, wh); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create webhook")
		return
	}

	// Return the secret only on creation (never again)
	type webhookResponse struct {
		*webhookTypes.Webhook
		Secret string `json:"secret"`
	}
	response.SendJSON(w, http.StatusCreated, webhookResponse{Webhook: wh, Secret: secret})
}

// ListWebhooks handles GET /api/v1/webhooks
func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := requestcontext.FromContext(ctx)

	webhooks, err := h.webhookRepo.ListByUserID(ctx, user.Id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to list webhooks")
		return
	}
	if webhooks == nil {
		webhooks = []*webhookTypes.Webhook{}
	}
	response.SendJSON(w, http.StatusOK, webhooks)
}

// DeleteWebhook handles DELETE /api/v1/webhooks/{id}
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := requestcontext.FromContext(ctx)
	id := mux.Vars(r)["id"]

	if err := h.webhookRepo.Delete(ctx, id, user.Id); err != nil {
		response.Error(w, http.StatusNotFound, "Webhook not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
