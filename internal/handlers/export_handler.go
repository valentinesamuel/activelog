package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	exportPkg "github.com/valentinesamuel/activelog/internal/export"
	"github.com/valentinesamuel/activelog/internal/jobs"
	queueTypes "github.com/valentinesamuel/activelog/internal/queue/types"
	"github.com/valentinesamuel/activelog/internal/repository"
	requestcontext "github.com/valentinesamuel/activelog/internal/requestContext"
	storageTypes "github.com/valentinesamuel/activelog/internal/storage/types"
	"github.com/valentinesamuel/activelog/pkg/response"
)

// ExportHandler handles activity export endpoints.
type ExportHandler struct {
	activityRepo  repository.ActivityRepositoryInterface
	exportRepo    *repository.ExportRepository
	queueProvider queueTypes.QueueProvider
	storage       storageTypes.StorageProvider
}

// ExportHandlerDeps contains the dependencies for ExportHandler.
type ExportHandlerDeps struct {
	ActivityRepo  repository.ActivityRepositoryInterface
	ExportRepo    *repository.ExportRepository
	QueueProvider queueTypes.QueueProvider
	Storage       storageTypes.StorageProvider
}

// NewExportHandler creates a new ExportHandler with the given dependencies.
func NewExportHandler(deps ExportHandlerDeps) *ExportHandler {
	return &ExportHandler{
		activityRepo:  deps.ActivityRepo,
		exportRepo:    deps.ExportRepo,
		queueProvider: deps.QueueProvider,
		storage:       deps.Storage,
	}
}

// ExportCSV streams the authenticated user's activities as a CSV download.
func (h *ExportHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := requestcontext.FromContext(ctx)

	activities, err := h.activityRepo.ListByUser(ctx, user.Id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="activities.csv"`)

	if err := exportPkg.ExportActivitiesCSV(ctx, activities, w); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate CSV export")
		return
	}
}

// EnqueuePDFExport creates a pending export record and enqueues a PDF generation job.
func (h *ExportHandler) EnqueuePDFExport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := requestcontext.FromContext(ctx)

	// Create export record
	record := &exportPkg.ExportRecord{
		UserID: user.Id,
		Format: exportPkg.FormatPDF,
		Status: exportPkg.StatusPending,
	}
	if err := h.exportRepo.Create(ctx, record); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create export record")
		return
	}

	// Marshal the job payload data
	payload := jobs.ExportPayload{
		UserID: user.Id,
		Format: string(exportPkg.FormatPDF),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to marshal job payload")
		return
	}

	// Enqueue the job
	jobPayload := queueTypes.JobPayload{
		Event: queueTypes.EventGenerateExport,
		Data:  data,
	}
	if _, err := h.queueProvider.Enqueue(ctx, queueTypes.InboxQueue, jobPayload); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to enqueue export job")
		return
	}

	response.SendJSON(w, http.StatusAccepted, map[string]string{
		"job_id": record.ID,
	})
}

// GetJobStatus returns the current status of an export job.
func (h *ExportHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	record, err := h.exportRepo.GetByID(ctx, jobID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Export job not found")
		return
	}

	response.SendJSON(w, http.StatusOK, record)
}

// GetDownloadURL generates a presigned URL for a completed export.
func (h *ExportHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	record, err := h.exportRepo.GetByID(ctx, jobID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Export job not found")
		return
	}

	if record.Status != exportPkg.StatusCompleted {
		response.Error(w, http.StatusConflict, "Export is not yet completed")
		return
	}

	if record.S3Key == nil {
		response.Error(w, http.StatusInternalServerError, "Export file key is missing")
		return
	}

	url, err := h.storage.GetPresignedURL(ctx, &storageTypes.PresignedURLInput{
		Key:       *record.S3Key,
		ExpiresIn: 15 * time.Minute,
		Operation: storageTypes.PresignGet,
	})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate download URL")
		return
	}

	response.SendJSON(w, http.StatusOK, map[string]string{
		"download_url": url,
	})
}
