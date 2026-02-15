# MONTH 6: Background Jobs & Email

**Weeks:** 21-24
**Phase:** Asynchronous Processing
**Theme:** Handle long-running tasks efficiently

---

## Overview

This month introduces asynchronous processing to your application. You'll learn how to offload long-running tasks to background workers, send emails, schedule recurring tasks, and generate reports. By the end, your application can handle complex operations without blocking user requests.

---

## API Endpoints Reference (for Postman Testing)

### Background Job Endpoints (Week 21)

**Trigger Weekly Summary Email:**
- **HTTP Method:** `POST`
- **URL:** `/api/v1/users/me/reports/weekly-summary`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (202 Accepted):**
  ```json
  {
    "message": "weekly summary email queued",
    "job_id": "d290f1ee-6c54-4b01-90e6-d701748f0851",
    "estimated_delivery": "2024-01-15T14:35:22Z"
  }
  ```

### Export Endpoints (Week 24)

**Request CSV Export:**
- **HTTP Method:** `POST`
- **URL:** `/api/v1/activities/export/csv`
- **Headers:**
  ```
  Content-Type: application/json
  Authorization: Bearer <your-jwt-token>
  ```
- **Request Body:**
  ```json
  {
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z",
    "include_tags": true
  }
  ```
- **Success Response (202 Accepted):**
  ```json
  {
    "message": "export job started",
    "job_id": "a1b2c3d4-5e6f-7g8h-9i0j-k1l2m3n4o5p6",
    "status_url": "/api/v1/jobs/a1b2c3d4-5e6f-7g8h-9i0j-k1l2m3n4o5p6",
    "estimated_completion": "2024-01-15T14:32:00Z"
  }
  ```

**Request PDF Report:**
- **HTTP Method:** `POST`
- **URL:** `/api/v1/activities/export/pdf`
- **Headers:**
  ```
  Content-Type: application/json
  Authorization: Bearer <your-jwt-token>
  ```
- **Request Body:**
  ```json
  {
    "report_type": "monthly",
    "month": "2024-01",
    "include_charts": true,
    "include_photos": false
  }
  ```
- **Success Response (202 Accepted):**
  ```json
  {
    "message": "pdf generation started",
    "job_id": "f1e2d3c4-b5a6-7890-1234-567890abcdef",
    "status_url": "/api/v1/jobs/f1e2d3c4-b5a6-7890-1234-567890abcdef",
    "estimated_completion": "2024-01-15T14:33:30Z"
  }
  ```

**Check Job Status:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/jobs/{job_id}`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (200 OK) - In Progress:**
  ```json
  {
    "job_id": "a1b2c3d4-5e6f-7g8h-9i0j-k1l2m3n4o5p6",
    "status": "processing",
    "progress": 45,
    "message": "Processing activities...",
    "created_at": "2024-01-15T14:30:00Z"
  }
  ```
- **Success Response (200 OK) - Completed:**
  ```json
  {
    "job_id": "a1b2c3d4-5e6f-7g8h-9i0j-k1l2m3n4o5p6",
    "status": "completed",
    "progress": 100,
    "result": {
      "download_url": "https://bucket.s3.amazonaws.com/exports/user-1-2024-01-activities.csv",
      "expires_at": "2024-01-22T14:30:00Z",
      "file_size": 245760,
      "record_count": 150
    },
    "created_at": "2024-01-15T14:30:00Z",
    "completed_at": "2024-01-15T14:31:45Z"
  }
  ```
- **Success Response (200 OK) - Failed:**
  ```json
  {
    "job_id": "a1b2c3d4-5e6f-7g8h-9i0j-k1l2m3n4o5p6",
    "status": "failed",
    "progress": 0,
    "error": "failed to generate export: database connection lost",
    "retry_count": 3,
    "created_at": "2024-01-15T14:30:00Z",
    "failed_at": "2024-01-15T14:32:15Z"
  }
  ```

**Download Export (Direct CSV Stream):**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/activities/export/csv/download`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Query Parameters:**
  ```
  ?start_date=2024-01-01&end_date=2024-01-31
  ```
- **Success Response (200 OK):**
  ```
  Headers:
    Content-Type: text/csv
    Content-Disposition: attachment; filename="activities-2024-01.csv"

  Body (CSV):
  id,activity_type,duration_minutes,distance_km,activity_date,tags,notes
  123,running,45,7.5,2024-01-15T06:30:00Z,"morning,outdoor,cardio","Morning run"
  122,yoga,30,0,2024-01-14T18:00:00Z,"evening,flexibility","Evening yoga"
  ...
  ```

### Email Verification Endpoints (Week 22)

**Resend Verification Email:**
- **HTTP Method:** `POST`
- **URL:** `/api/v1/auth/resend-verification`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (202 Accepted):**
  ```json
  {
    "message": "verification email queued",
    "email": "john@example.com"
  }
  ```

**Verify Email:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/auth/verify-email?token=abc123def456`
- **Query Parameters:**
  ```
  token=abc123def456
  ```
- **Success Response (200 OK):**
  ```json
  {
    "message": "email verified successfully",
    "verified_at": "2024-01-15T14:30:22Z"
  }
  ```
- **Error Response (400 Bad Request):**
  ```json
  {
    "error": "invalid token",
    "message": "verification token is invalid or expired"
  }
  ```

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

# WEEKLY TASK BREAKDOWNS

## Week 21: Job Queue System

### 📋 Implementation Tasks

**Task 1: Install and Configure Asynq** (30 min)
- [ ] Install Asynq: `go get github.com/hibiken/asynq`
- [ ] Ensure Redis is running (asynq uses Redis as queue backend)
- [ ] Test basic asynq client/server setup
- [ ] Review asynq documentation and examples

**Task 2: Define Job Types and Payloads** (45 min)
- [ ] Create `internal/jobs/types.go`
- [ ] Define job type constants: `TypeWelcomeEmail`, `TypeWeeklySummary`, etc.
- [ ] Create payload structs for each job type
- [ ] Add JSON tags for serialization
- [ ] Document expected payload structure

**Task 3: Create Job Client** (60 min)
- [ ] Create `internal/jobs/client.go`
- [ ] Implement `NewJobClient(redisAddr) (*JobClient, error)`
  - **Logic:** Create `asynq.RedisClientOpt` with Redis address. Create `asynq.Client` with options. Return JobClient struct wrapping the asynq client. Used by API server to enqueue jobs.
- [ ] Add method `EnqueueWelcomeEmail(ctx, userID, email, name) error`
  - **Logic:** Create payload struct with userID, email, name. Marshal to JSON. Create `asynq.NewTask(TypeWelcomeEmail, payload)`. Call `client.Enqueue(task, asynq.MaxRetry(3), asynq.Queue("critical"))`. Returns task ID on success.
- [ ] Add method `EnqueueWeeklySummary(ctx, userID) error`
  - **Logic:** Similar to welcome email but payload only has userID. Use different queue priority (default queue). Set `asynq.ProcessIn(5*time.Minute)` to delay processing.
- [ ] Set appropriate retry policies (max 3 retries)
- [ ] Set timeouts for different job types

**Task 4: Implement Job Handlers** (90 min)
- [ ] Create `internal/jobs/handlers.go`
- [ ] Implement `HandleWelcomeEmail(ctx, task) error`
  - **Logic:** Unmarshal task.Payload() into WelcomeEmailPayload struct. Call emailService.SendWelcomeEmail(payload.Email, payload.Name). If error, return it (asynq will retry). If success, return nil. Log start and completion.
- [ ] Implement `HandleWeeklySummary(ctx, task) error`
  - **Logic:** Unmarshal payload to get userID. Fetch user's weekly stats from repository. Fetch user email. Render email template with stats data. Send email. Return error for retry if any step fails.
- [ ] Implement `HandleGenerateReport(ctx, task) error`
  - **Logic:** Unmarshal payload. Generate PDF report using activity data. Upload PDF to S3. Store S3 key in database. Send notification email with download link. Can take 30+ seconds for large reports.
- [ ] Unmarshal task payload in each handler
- [ ] Add error handling and logging
- [ ] Return errors for retry on failure

**Task 5: Create Worker Server** (60 min)
- [ ] Create `cmd/worker/main.go`
- [ ] Initialize asynq server with Redis connection
- [ ] Configure concurrency (10 workers)
- [ ] Set up queue priorities (critical: 6, default: 3, low: 1)
- [ ] Register job handlers with mux
- [ ] Add graceful shutdown handling
- [ ] Test worker processes jobs

**Task 6: Integrate Job Enqueueing** (45 min)
- [ ] Update user registration to enqueue welcome email
- [ ] Update activity creation to enqueue notifications
- [ ] Add job client to service dependencies
- [ ] Test job enqueueing from API endpoints
- [ ] Verify jobs appear in Redis (use `redis-cli`)

**Task 7: Add Job Monitoring** (30 min)
- [ ] Install asynq web UI (optional): `go get github.com/hibiken/asynq/tools/asynq`
- [ ] Add job metrics to Prometheus
- [ ] Track jobs enqueued, processed, failed
- [ ] Monitor queue lengths
- [ ] Set up alerts for job failures

### 📦 Files You'll Create/Modify

```
internal/
├── jobs/
│   ├── types.go                   [CREATE]
│   ├── client.go                  [CREATE]
│   ├── handlers.go                [CREATE]
│   └── handlers_test.go           [CREATE]

cmd/
├── worker/
│   └── main.go                    [CREATE]
└── api/
    └── main.go                    [MODIFY - add job client]

Makefile                           [MODIFY - add worker target]
```

### 🔄 Implementation Order

1. **Setup**: Install asynq → Test Redis connection
2. **Types**: Define job types and payloads
3. **Client**: Job client for enqueueing
4. **Handlers**: Implement job processing logic
5. **Worker**: Create worker server
6. **Integration**: Enqueue jobs from API
7. **Monitoring**: Add metrics and monitoring

### ⚠️ Blockers to Watch For

- **Redis dependency**: Worker crashes if Redis unavailable
- **Serialization**: JSON marshal/unmarshal errors - validate payloads
- **Concurrency**: Too many workers = resource exhaustion
- **Retries**: Infinite retry loops - set max retries
- **Dead letter queue**: Failed jobs need manual intervention
- **Memory leaks**: Worker goroutines must clean up properly

### ✅ Definition of Done

- [ ] Asynq client and server configured
- [ ] Can enqueue jobs from API
- [ ] Worker processes jobs successfully
- [ ] Failed jobs retry automatically (max 3 times)
- [ ] Job metrics tracked in Prometheus
- [ ] Worker runs as separate process
- [ ] Graceful shutdown working

---

## Week 22: Email Integration

### 📋 Implementation Tasks

**Task 1: Choose Email Provider** (20 min)
- [ ] Select provider: SendGrid, Mailgun, or SMTP
- [ ] Create account and get API keys/credentials
- [ ] For SMTP: get host, port, username, password
- [ ] Store credentials in `.env` (never commit!)
- [ ] Add to config struct

**Task 2: Install Email Library** (15 min)
- [ ] For SMTP: `go get gopkg.in/gomail.v2`
- [ ] Or for SendGrid: `go get github.com/sendgrid/sendgrid-go`
- [ ] Test imports work

**Task 3: Create Email Service** (60 min)
- [ ] Create `internal/email/service.go`
- [ ] Implement `NewEmailService(config) *EmailService`
  - **Logic:** Parse config for SMTP host, port, username, password. Create `gomail.Dialer` with TLS config. Parse all HTML templates from templates/ directory. Return EmailService with dialer and parsed templates.
- [ ] Add `SendEmail(to, subject, body) error` method
  - **Logic:** Create `gomail.NewMessage()`. Set From, To, Subject, HTML body. Call `dialer.DialAndSend(msg)`. Returns error if SMTP fails. Retry logic handled by asynq, not here.
- [ ] Configure SMTP dialer or API client
- [ ] Add connection pooling for SMTP
- [ ] Handle send errors gracefully

**Task 4: Create Email Templates** (90 min)
- [ ] Create `internal/email/templates/` directory
- [ ] Create `welcome.html` template
- [ ] Create `weekly_summary.html` template
- [ ] Create `friend_request.html` template
- [ ] Use Go's `html/template` package
- [ ] Add CSS for styling (inline styles)
- [ ] Test templates render correctly

**Task 5: Implement Template Rendering** (60 min)
- [ ] Create `RenderTemplate(name, data) (string, error)` function
- [ ] Parse template files on service initialization
- [ ] Execute templates with data
- [ ] Handle template errors
- [ ] Cache parsed templates for performance

**Task 6: Implement Specific Email Types** (90 min)
- [ ] Implement `SendWelcomeEmail(to, name) error`
  - **Logic:** Build template data with name. Execute welcome.html template with data. Get HTML string. Call SendEmail(to, "Welcome to ActiveLog!", html). Return error if template execution or send fails.
- [ ] Implement `SendWeeklySummary(userID) error`
  - **Logic:**
    1. Fetch user from repository to get email
    2. Fetch weekly stats from stats repository
    3. Build template data with user name, stats (total activities, distance, time, top activity type)
    4. Execute weekly_summary.html template
    5. Send email with rendered HTML
  - Fetch user's weekly stats from repository
  - Render with template
  - Send email
- [ ] Implement `SendFriendRequest(to, fromName) error`
  - **Logic:** Execute friend_request.html template with fromName data. Send email with "Friend Request from {fromName}" subject.
- [ ] Test each email type

**Task 7: Integrate with Job Queue** (45 min)
- [ ] Update `HandleWelcomeEmail` job to use email service
- [ ] Update `HandleWeeklySummary` job to use email service
- [ ] Pass email service to job handlers via dependency injection
- [ ] Test end-to-end: enqueue job → worker sends email
- [ ] Verify emails received

### 📦 Files You'll Create/Modify

```
internal/
├── email/
│   ├── service.go                 [CREATE]
│   ├── service_test.go            [CREATE]
│   ├── templates/
│   │   ├── welcome.html           [CREATE]
│   │   ├── weekly_summary.html    [CREATE]
│   │   └── friend_request.html    [CREATE]
│   └── template_data.go           [CREATE - template structs]

internal/jobs/
└── handlers.go                    [MODIFY - use email service]

.env                               [MODIFY - add email config]
```

### 🔄 Implementation Order

1. **Provider**: Choose and configure email provider
2. **Service**: Email service with send method
3. **Templates**: HTML email templates
4. **Rendering**: Template rendering logic
5. **Email types**: Specific email implementations
6. **Integration**: Connect to job queue
7. **Testing**: Send test emails

### ⚠️ Blockers to Watch For

- **Credentials**: Never commit SMTP/API credentials
- **Rate limits**: Email providers limit sends per day/hour
- **Spam filters**: Emails might go to spam - configure SPF/DKIM
- **Template errors**: Invalid templates crash at runtime
- **HTML rendering**: Email clients render HTML differently
- **Inline styles**: Use inline CSS, not external stylesheets

### ✅ Definition of Done

- [ ] Email service configured with provider
- [ ] Can send emails successfully
- [ ] HTML templates created and tested
- [ ] Welcome emails sent on user registration
- [ ] Weekly summary emails working
- [ ] Emails look good in Gmail/Outlook
- [ ] No credentials in git repository

---

## Week 23: Scheduled Tasks (Cron)

### 📋 Implementation Tasks

**Task 1: Install Cron Library** (10 min)
- [ ] Install: `go get github.com/robfig/cron/v3`
- [ ] Review cron syntax documentation
- [ ] Test basic cron scheduling

**Task 2: Create Scheduler Service** (45 min)
- [ ] Create `internal/scheduler/scheduler.go`
- [ ] Implement `NewScheduler(services) *Scheduler`
- [ ] Initialize cron instance
- [ ] Add `Start()` and `Stop()` methods
- [ ] Add graceful shutdown handling

**Task 3: Implement Daily Statistics Calculation** (90 min)
- [ ] Create `internal/services/stats_calculator.go`
- [ ] Implement `CalculateDailyStats(ctx) error`
- [ ] Aggregate previous day's activities
- [ ] Calculate totals per user
- [ ] Store in `daily_stats` table (create migration)
- [ ] Schedule to run at midnight: `"0 0 * * *"`

**Task 4: Implement Weekly Email Job** (60 min)
- [ ] Add schedule: `"0 9 * * 0"` (Sunday 9am)
- [ ] Fetch all active users
- [ ] For each user, enqueue weekly summary email job
- [ ] Add throttling to avoid overwhelming job queue
- [ ] Log how many emails queued

**Task 5: Implement Monthly Report Generation** (75 min)
- [ ] Add schedule: `"0 0 1 * *"` (1st of month, midnight)
- [ ] Generate PDF reports for all users
- [ ] Store reports in S3: `reports/{userID}/{year}-{month}.pdf`
- [ ] Enqueue notification emails with report link
- [ ] Test report generation

**Task 6: Implement Cleanup Jobs** (60 min)
- [ ] Old deleted activities cleanup: `"0 2 * * *"` (daily 2am)
- [ ] Permanently delete soft-deleted records older than 30 days
- [ ] Cleanup orphaned photos in S3
- [ ] Cleanup expired cache entries
- [ ] Log cleanup statistics

**Task 7: Add Scheduler Monitoring** (30 min)
- [ ] Track cron job executions in Prometheus
- [ ] Track execution duration
- [ ] Track failures
- [ ] Alert on consecutive failures
- [ ] Add to Grafana dashboard

### 📦 Files You'll Create/Modify

```
migrations/
├── 007_create_daily_stats.up.sql  [CREATE]
└── 007_create_daily_stats.down.sql [CREATE]

internal/
├── scheduler/
│   ├── scheduler.go               [CREATE]
│   └── scheduler_test.go          [CREATE]
├── services/
│   ├── stats_calculator.go        [CREATE]
│   └── cleanup_service.go         [CREATE]

cmd/api/
└── main.go                        [MODIFY - start scheduler]
```

### 🔄 Implementation Order

1. **Setup**: Install cron library → Create scheduler
2. **Daily stats**: Calculate and store daily aggregates
3. **Weekly emails**: Schedule weekly summaries
4. **Monthly reports**: Generate PDF reports
5. **Cleanup**: Automated cleanup jobs
6. **Monitoring**: Add metrics and alerts

### ⚠️ Blockers to Watch For

- **Timezone**: Cron runs in server timezone - use UTC
- **Long-running jobs**: Don't block cron scheduler
- **Concurrent runs**: Prevent job overlap (use mutex or skip if running)
- **Failures**: Log failures, don't crash entire scheduler
- **Testing**: Hard to test time-based jobs - use dependency injection
- **DST**: Daylight saving time can affect schedules

### ✅ Definition of Done

- [ ] Daily stats calculated at midnight
- [ ] Weekly emails sent every Sunday at 9am
- [ ] Monthly reports generated on 1st of month
- [ ] Cleanup jobs running daily at 2am
- [ ] All schedules in UTC timezone
- [ ] Cron metrics tracked
- [ ] Can manually trigger jobs for testing

---

## Week 24: Export Features (PDF/CSV)

### 📋 Implementation Tasks

**Task 1: Install PDF Library** (15 min)
- [ ] Install: `go get github.com/jung-kurt/gofpdf`
- [ ] Or alternative: `go get github.com/johnfercher/maroto/v2`
- [ ] Test basic PDF generation

**Task 2: Implement CSV Export** (60 min)
- [ ] Create `internal/export/csv_exporter.go`
- [ ] Implement `ExportActivities(ctx, userID, writer) error`
- [ ] Write CSV header row
- [ ] Stream activities from database
- [ ] Write each row to CSV
- [ ] Handle large datasets (streaming, not loading all in memory)

**Task 3: Create CSV Download Endpoint** (45 min)
- [ ] Add `GET /api/v1/export/activities/csv` endpoint
- [ ] Set response headers: `Content-Type: text/csv`
- [ ] Set `Content-Disposition: attachment; filename=activities.csv`
- [ ] Stream CSV directly to response
- [ ] Test download in browser

**Task 4: Implement PDF Report Generation** (120 min)
- [ ] Create `internal/export/pdf_exporter.go`
- [ ] Implement `GenerateActivityReport(ctx, userID) ([]byte, error)`
- [ ] Add report title and user info
- [ ] Add summary statistics section
- [ ] Add activity table with pagination
- [ ] Add charts (optional, using chart library)
- [ ] Style with colors and fonts

**Task 5: Create Async Export System** (90 min)
- [ ] Create background job for large exports
- [ ] Store export file in S3: `exports/{userID}/{export_id}.pdf`
- [ ] Create `exports` table to track export status
- [ ] Add endpoint: `POST /api/v1/export/request` (returns job ID)
- [ ] Add endpoint: `GET /api/v1/export/{id}/status`
- [ ] Add endpoint: `GET /api/v1/export/{id}/download`
- [ ] Send email when export ready

**Task 6: Implement Export Job Handler** (60 min)
- [ ] Add `HandleGenerateExport` job handler
- [ ] Support both CSV and PDF formats
- [ ] Upload result to S3
- [ ] Update export record status
- [ ] Send completion email with download link
- [ ] Set S3 file expiration (7 days)

**Task 7: Add Export Management UI** (45 min)
- [ ] Add endpoint: `GET /api/v1/export/history` (list user's exports)
- [ ] Add endpoint: `DELETE /api/v1/export/{id}` (delete export)
- [ ] Show export status (pending, processing, completed, failed)
- [ ] Show download link when ready
- [ ] Show expiration date

### 📦 Files You'll Create/Modify

```
migrations/
├── 008_create_exports.up.sql      [CREATE]
└── 008_create_exports.down.sql    [CREATE]

internal/
├── export/
│   ├── csv_exporter.go            [CREATE]
│   ├── pdf_exporter.go            [CREATE]
│   ├── exporter_test.go           [CREATE]
│   └── types.go                   [CREATE]
├── handlers/
│   └── export_handler.go          [CREATE]
├── jobs/
│   └── handlers.go                [MODIFY - add export handler]
└── repository/
    └── export_repository.go       [CREATE]
```

### 🔄 Implementation Order

1. **CSV**: Simple CSV export → Download endpoint
2. **PDF**: PDF generation → Styling
3. **Async**: Background job system → S3 upload
4. **Job handler**: Export job processing
5. **Management**: List/delete exports
6. **Testing**: Test large exports don't timeout

### ⚠️ Blockers to Watch For

- **Memory**: Don't load entire dataset - stream rows
- **Timeout**: Large exports can timeout - use background jobs
- **S3 costs**: Exports consume storage - set expiration
- **PDF size**: Large reports can be slow - paginate or limit
- **Concurrent exports**: Limit concurrent exports per user
- **Cleanup**: Delete old exports from S3

### ✅ Definition of Done

- [ ] Can download CSV of activities instantly
- [ ] Can generate PDF report with stats and charts
- [ ] Large exports processed in background
- [ ] Email sent when export ready
- [ ] Can download export from S3 link
- [ ] Exports expire after 7 days
- [ ] Can view export history
- [ ] All tests passing

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
