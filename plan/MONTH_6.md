# MONTH 6: Background Jobs & Email

**Weeks:** 21-24
**Phase:** Asynchronous Processing
**Theme:** Handle long-running tasks efficiently

---

## Overview

This month introduces asynchronous processing to your application. You'll learn how to offload long-running tasks to background workers, send emails, schedule recurring tasks, and generate reports. By the end, your application can handle complex operations without blocking user requests.

---

## Learning Path

### Week 21: Job Queue System
- Understand worker pool pattern
- Implement job queue with Redis
- Create background workers
- Handle job failures and retries

### Week 22: Email Integration
- SMTP configuration
- Email templates
- Send welcome emails
- Transactional emails

### Week 23: Scheduled Tasks (Cron)
- Cron job syntax
- Implement scheduled tasks in Go
- Daily/weekly statistics
- Cleanup old data

### Week 24: Export Features (PDF/CSV)
- Generate CSV exports
- Create PDF reports
- Stream large exports
- Background export jobs

---

## Worker Pool Pattern

```go
import "github.com/hibiken/asynq"

// Define job types
const (
    TypeWelcomeEmail   = "email:welcome"
    TypeWeeklySummary  = "email:weekly_summary"
    TypeGenerateReport = "report:generate"
)

// Job payload
type EmailPayload struct {
    UserID int    `json:"user_id"`
    Email  string `json:"email"`
    Name   string `json:"name"`
}

// Enqueue job
func (s *JobService) EnqueueWelcomeEmail(ctx context.Context, userID int, email, name string) error {
    payload, err := json.Marshal(EmailPayload{
        UserID: userID,
        Email:  email,
        Name:   name,
    })
    if err != nil {
        return err
    }

    task := asynq.NewTask(TypeWelcomeEmail, payload)

    // Enqueue with options
    _, err = s.client.EnqueueContext(ctx, task,
        asynq.MaxRetry(3),
        asynq.Timeout(5*time.Minute),
    )

    return err
}

// Worker to process jobs
func HandleWelcomeEmail(ctx context.Context, t *asynq.Task) error {
    var payload EmailPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("unmarshal payload: %w", err)
    }

    // Send email
    emailService := ctx.Value("emailService").(*EmailService)
    return emailService.SendWelcomeEmail(payload.Email, payload.Name)
}

// Start workers
func StartWorkers(redisAddr string) *asynq.Server {
    srv := asynq.NewServer(
        asynq.RedisClientOpt{Addr: redisAddr},
        asynq.Config{
            Concurrency: 10, // 10 concurrent workers
            Queues: map[string]int{
                "critical": 6, // 60% of workers
                "default":  3, // 30% of workers
                "low":      1, // 10% of workers
            },
        },
    )

    mux := asynq.NewServeMux()
    mux.HandleFunc(TypeWelcomeEmail, HandleWelcomeEmail)
    mux.HandleFunc(TypeWeeklySummary, HandleWeeklySummary)
    mux.HandleFunc(TypeGenerateReport, HandleGenerateReport)

    return srv
}
```

---

## Email Features

### SMTP Configuration
```go
import "gopkg.in/gomail.v2"

type EmailService struct {
    dialer *gomail.Dialer
    from   string
}

func NewEmailService(host string, port int, username, password string) *EmailService {
    return &EmailService{
        dialer: gomail.NewDialer(host, port, username, password),
        from:   username,
    }
}

func (s *EmailService) SendEmail(to, subject, body string) error {
    m := gomail.NewMessage()
    m.SetHeader("From", s.from)
    m.SetHeader("To", to)
    m.SetHeader("Subject", subject)
    m.SetBody("text/html", body)

    return s.dialer.DialAndSend(m)
}
```

### Email Templates
```go
import "html/template"

const welcomeEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .button { background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Welcome to ActiveLog, {{.Name}}!</h1>
        <p>Thank you for joining our community of active individuals.</p>
        <p>Start tracking your activities today:</p>
        <a href="{{.AppURL}}/dashboard" class="button">Go to Dashboard</a>
    </div>
</body>
</html>
`

type WelcomeEmailData struct {
    Name   string
    AppURL string
}

func (s *EmailService) SendWelcomeEmail(to, name string) error {
    tmpl, err := template.New("welcome").Parse(welcomeEmailTemplate)
    if err != nil {
        return err
    }

    var body bytes.Buffer
    data := WelcomeEmailData{
        Name:   name,
        AppURL: os.Getenv("APP_URL"),
    }

    if err := tmpl.Execute(&body, data); err != nil {
        return err
    }

    return s.SendEmail(to, "Welcome to ActiveLog!", body.String())
}
```

### Weekly Summary Email
```go
func (s *EmailService) SendWeeklySummary(userID int) error {
    // Get user stats for the week
    stats, err := s.statsRepo.GetWeeklyStats(context.Background(), userID)
    if err != nil {
        return err
    }

    user, err := s.userRepo.GetByID(context.Background(), userID)
    if err != nil {
        return err
    }

    // Generate email body
    body := fmt.Sprintf(`
        <h2>Your Week in Review</h2>
        <p>Hi %s,</p>
        <ul>
            <li>Activities completed: %d</li>
            <li>Total distance: %.2f km</li>
            <li>Total duration: %d minutes</li>
            <li>Calories burned: %d</li>
        </ul>
        <p>Keep up the great work!</p>
    `, user.Name, stats.Activities, stats.Distance, stats.Duration, stats.Calories)

    return s.SendEmail(user.Email, "Your Weekly Summary", body)
}
```

---

## Scheduled Tasks

### Cron Implementation
```go
import "github.com/robfig/cron/v3"

func StartScheduledTasks(services *Services) *cron.Cron {
    c := cron.New()

    // Daily statistics calculation (midnight)
    c.AddFunc("0 0 * * *", func() {
        log.Println("Running daily statistics calculation...")
        if err := services.Stats.CalculateDailyStats(context.Background()); err != nil {
            log.Printf("Daily stats error: %v", err)
        }
    })

    // Weekly email summaries (Sunday 9am)
    c.AddFunc("0 9 * * 0", func() {
        log.Println("Sending weekly summaries...")
        if err := services.Email.SendWeeklySummaries(context.Background()); err != nil {
            log.Printf("Weekly summaries error: %v", err)
        }
    })

    // Monthly report generation (1st of month, midnight)
    c.AddFunc("0 0 1 * *", func() {
        log.Println("Generating monthly reports...")
        if err := services.Reports.GenerateMonthlyReports(context.Background()); err != nil {
            log.Printf("Monthly reports error: %v", err)
        }
    })

    // Cleanup old data (daily at 2am)
    c.AddFunc("0 2 * * *", func() {
        log.Println("Cleaning up old data...")
        if err := services.Cleanup.DeleteOldData(context.Background()); err != nil {
            log.Printf("Cleanup error: %v", err)
        }
    })

    c.Start()
    return c
}

// Cron syntax reminder:
// ┌───────────── minute (0 - 59)
// │ ┌───────────── hour (0 - 23)
// │ │ ┌───────────── day of month (1 - 31)
// │ │ │ ┌───────────── month (1 - 12)
// │ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
// │ │ │ │ │
// * * * * *
```

---

## Export Features

### CSV Export
```go
import "encoding/csv"

func (s *ExportService) ExportActivitiesCSV(ctx context.Context, userID int, w io.Writer) error {
    activities, err := s.repo.GetByUserID(ctx, userID)
    if err != nil {
        return err
    }

    writer := csv.NewWriter(w)
    defer writer.Flush()

    // Write header
    header := []string{"Date", "Type", "Duration (min)", "Distance (km)", "Calories", "Notes"}
    if err := writer.Write(header); err != nil {
        return err
    }

    // Write data
    for _, activity := range activities {
        row := []string{
            activity.Date.Format("2006-01-02"),
            activity.Type,
            strconv.Itoa(activity.Duration),
            fmt.Sprintf("%.2f", activity.Distance),
            strconv.Itoa(activity.Calories),
            activity.Notes,
        }
        if err := writer.Write(row); err != nil {
            return err
        }
    }

    return nil
}

// Handler for CSV download
func (h *ExportHandler) DownloadCSV(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())

    // Set headers for download
    w.Header().Set("Content-Type", "text/csv")
    w.Header().Set("Content-Disposition", "attachment; filename=activities.csv")

    if err := h.service.ExportActivitiesCSV(r.Context(), userID, w); err != nil {
        http.Error(w, "Export failed", http.StatusInternalServerError)
        return
    }
}
```

### PDF Report Generation
```go
import "github.com/jung-kurt/gofpdf"

func (s *ReportService) GeneratePDFReport(ctx context.Context, userID int) ([]byte, error) {
    // Get user data
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    stats, err := s.statsRepo.GetMonthlyStats(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Create PDF
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    // Title
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(40, 10, "Monthly Activity Report")
    pdf.Ln(12)

    // User info
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(40, 10, fmt.Sprintf("User: %s", user.Name))
    pdf.Ln(8)
    pdf.Cell(40, 10, fmt.Sprintf("Period: %s", time.Now().Format("January 2006")))
    pdf.Ln(12)

    // Statistics
    pdf.SetFont("Arial", "B", 14)
    pdf.Cell(40, 10, "Summary Statistics")
    pdf.Ln(8)

    pdf.SetFont("Arial", "", 12)
    pdf.Cell(40, 10, fmt.Sprintf("Total Activities: %d", stats.TotalActivities))
    pdf.Ln(6)
    pdf.Cell(40, 10, fmt.Sprintf("Total Distance: %.2f km", stats.TotalDistance))
    pdf.Ln(6)
    pdf.Cell(40, 10, fmt.Sprintf("Total Duration: %d minutes", stats.TotalDuration))
    pdf.Ln(6)

    // Convert to bytes
    var buf bytes.Buffer
    if err := pdf.Output(&buf); err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}
```

### Background Export Job
```go
// Enqueue export job for large datasets
func (h *ExportHandler) RequestExport(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())
    format := r.URL.Query().Get("format") // csv or pdf

    // Create export job
    payload := ExportPayload{
        UserID: userID,
        Format: format,
    }

    data, _ := json.Marshal(payload)
    task := asynq.NewTask(TypeGenerateExport, data)

    info, err := h.jobClient.Enqueue(task, asynq.Queue("low"))
    if err != nil {
        response.Error(w, http.StatusInternalServerError, "Failed to queue export")
        return
    }

    // Return job ID for polling
    response.JSON(w, http.StatusAccepted, map[string]interface{}{
        "job_id": info.ID,
        "status": "queued",
        "message": "Export is being generated. Check status with job_id.",
    })
}

// Handler to check export status
func (h *ExportHandler) GetExportStatus(w http.ResponseWriter, r *http.Request) {
    jobID := chi.URLParam(r, "jobId")

    // Check job status in Redis
    status, err := h.jobClient.GetTaskInfo(jobID)
    if err != nil {
        response.Error(w, http.StatusNotFound, "Job not found")
        return
    }

    response.JSON(w, http.StatusOK, map[string]interface{}{
        "job_id": jobID,
        "status": status.State,
        "download_url": fmt.Sprintf("/api/v1/exports/%s/download", jobID),
    })
}
```

---

## Common Pitfalls

1. **Blocking user requests with long tasks**
   - ❌ Generating reports in HTTP handler
   - ✅ Use background jobs

2. **No job retry logic**
   - ❌ Jobs fail permanently
   - ✅ Implement exponential backoff retries

3. **Sending emails synchronously**
   - ❌ Slow API responses
   - ✅ Queue email jobs

4. **No email rate limiting**
   - ❌ Spam users, get blacklisted
   - ✅ Limit email frequency

5. **Hardcoded email templates**
   - ❌ Can't update without redeployment
   - ✅ Store templates in database or files

---

## Testing

```go
func TestWeeklySummaryJob(t *testing.T) {
    // Use asynq test utilities
    ctx := context.Background()

    payload := EmailPayload{
        UserID: 1,
        Email:  "test@example.com",
    }

    data, _ := json.Marshal(payload)
    task := asynq.NewTask(TypeWeeklySummary, data)

    err := HandleWeeklySummary(ctx, task)
    assert.NoError(t, err)
}
```

---

## Resources

- [asynq - Job Queue Library](https://github.com/hibiken/asynq)
- [robfig/cron - Cron Library](https://github.com/robfig/cron)
- [gomail - Email Library](https://github.com/go-gomail/gomail)
- [gofpdf - PDF Generation](https://github.com/jung-kurt/gofpdf)

---

## Next Steps

After completing Month 6, you'll move to **Month 7: Concurrency Deep Dive**, where you'll learn:
- Goroutines fundamentals
- Channels and select statements
- Sync primitives
- Race detection
- Concurrency patterns

**Your API can now handle complex async operations!** 🚀
