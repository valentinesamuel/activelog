# MONTH 9: Analytics & Reporting

**Weeks:** 33-36
**Phase:** Data Analysis & Insights
**Theme:** Turn data into actionable insights

---

## Overview

This month focuses on extracting insights from your activity data. You'll write advanced SQL queries, implement a goal tracking system, generate reports, and create chart data endpoints. By the end, users will have powerful analytics to track their progress and achieve their fitness goals.

---

## API Endpoints Reference (for Postman Testing)

### Advanced Analytics Endpoints (Week 33)

**Get Monthly Trends:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/analytics/trends/monthly`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Query Parameters:**
  ```
  ?months=12&activity_type=running
  ```
- **Success Response (200 OK):**
  ```json
  {
    "trends": [
      {
        "month": "2024-01",
        "total_activities": 18,
        "total_distance_km": 95.5,
        "total_duration_minutes": 540,
        "average_pace_min_per_km": 5.65,
        "vs_previous_month": {
          "activities_change_pct": 12.5,
          "distance_change_pct": 8.3
        }
      },
      {
        "month": "2024-02",
        "total_activities": 20,
        "total_distance_km": 108.2,
        "total_duration_minutes": 600,
        "average_pace_min_per_km": 5.54,
        "vs_previous_month": {
          "activities_change_pct": 11.1,
          "distance_change_pct": 13.3
        }
      }
    ],
    "total_months": 12
  }
  ```

**Get Personal Records:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/analytics/records`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (200 OK):**
  ```json
  {
    "records": {
      "longest_run_km": {
        "value": 21.1,
        "activity_id": 456,
        "activity_date": "2024-06-15T06:00:00Z"
      },
      "longest_duration_minutes": {
        "value": 135,
        "activity_id": 457,
        "activity_date": "2024-07-20T06:00:00Z"
      },
      "fastest_5k_pace": {
        "value": 4.45,
        "activity_id": 458,
        "activity_date": "2024-08-10T06:00:00Z"
      },
      "longest_streak_days": {
        "value": 45,
        "start_date": "2024-01-01T00:00:00Z",
        "end_date": "2024-02-15T00:00:00Z"
      }
    }
  }
  ```

### Goal Tracking Endpoints (Week 34)

**Create Goal:**
- **HTTP Method:** `POST`
- **URL:** `/api/v1/goals`
- **Headers:**
  ```
  Content-Type: application/json
  Authorization: Bearer <your-jwt-token>
  ```
- **Request Body:**
  ```json
  {
    "goal_type": "distance",
    "target_value": 200,
    "activity_type": "running",
    "period": "monthly",
    "start_date": "2024-02-01T00:00:00Z"
  }
  ```
- **Success Response (201 Created):**
  ```json
  {
    "id": 10,
    "goal_type": "distance",
    "target_value": 200,
    "activity_type": "running",
    "period": "monthly",
    "start_date": "2024-02-01T00:00:00Z",
    "end_date": "2024-02-29T23:59:59Z",
    "status": "active",
    "created_at": "2024-01-15T14:30:22Z"
  }
  ```

**Get Active Goals with Progress:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/goals/active`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (200 OK):**
  ```json
  {
    "goals": [
      {
        "id": 10,
        "goal_type": "distance",
        "target_value": 200,
        "current_value": 145.5,
        "progress_pct": 72.75,
        "activity_type": "running",
        "period": "monthly",
        "start_date": "2024-02-01T00:00:00Z",
        "end_date": "2024-02-29T23:59:59Z",
        "days_remaining": 14,
        "on_track": true,
        "status": "active"
      },
      {
        "id": 11,
        "goal_type": "frequency",
        "target_value": 20,
        "current_value": 12,
        "progress_pct": 60.0,
        "activity_type": "any",
        "period": "monthly",
        "start_date": "2024-02-01T00:00:00Z",
        "end_date": "2024-02-29T23:59:59Z",
        "days_remaining": 14,
        "on_track": false,
        "status": "active"
      }
    ],
    "total": 2
  }
  ```

**Update Goal:**
- **HTTP Method:** `PATCH`
- **URL:** `/api/v1/goals/{id}`
- **Headers:**
  ```
  Content-Type: application/json
  Authorization: Bearer <your-jwt-token>
  ```
- **Request Body:**
  ```json
  {
    "target_value": 250,
    "status": "active"
  }
  ```
- **Success Response (200 OK):**
  ```json
  {
    "id": 10,
    "target_value": 250,
    "status": "active",
    "updated_at": "2024-01-15T14:35:22Z"
  }
  ```

**Delete Goal:**
- **HTTP Method:** `DELETE`
- **URL:** `/api/v1/goals/{id}`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (204 No Content):**
  ```
  (empty body)
  ```

### Chart Data Endpoints (Week 36)

**Get Activity Distribution Chart Data:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/analytics/charts/distribution`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Query Parameters:**
  ```
  ?period=last_30_days
  ```
- **Success Response (200 OK):**
  ```json
  {
    "chart_type": "pie",
    "data": [
      {
        "label": "Running",
        "value": 45,
        "percentage": 45.0,
        "color": "#FF6384"
      },
      {
        "label": "Cycling",
        "value": 25,
        "percentage": 25.0,
        "color": "#36A2EB"
      },
      {
        "label": "Yoga",
        "value": 15,
        "percentage": 15.0,
        "color": "#FFCE56"
      },
      {
        "label": "Swimming",
        "value": 10,
        "percentage": 10.0,
        "color": "#4BC0C0"
      },
      {
        "label": "Other",
        "value": 5,
        "percentage": 5.0,
        "color": "#9966FF"
      }
    ],
    "total": 100
  }
  ```

**Get Time Series Chart Data:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/analytics/charts/timeseries`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Query Parameters:**
  ```
  ?metric=distance&granularity=daily&days=30&activity_type=running
  ```
- **Success Response (200 OK):**
  ```json
  {
    "chart_type": "line",
    "metric": "distance_km",
    "granularity": "daily",
    "data_points": [
      {
        "date": "2024-01-01",
        "value": 5.2
      },
      {
        "date": "2024-01-02",
        "value": 0
      },
      {
        "date": "2024-01-03",
        "value": 7.8
      },
      {
        "date": "2024-01-04",
        "value": 6.5
      }
    ],
    "summary": {
      "total": 185.5,
      "average": 6.18,
      "max": 12.5,
      "min": 0
    }
  }
  ```

**Get Comparison Chart Data:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/analytics/charts/comparison`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Query Parameters:**
  ```
  ?metric=duration&period=monthly&compare_periods=3
  ```
- **Success Response (200 OK):**
  ```json
  {
    "chart_type": "bar",
    "metric": "duration_minutes",
    "data": [
      {
        "period": "December 2024",
        "value": 540,
        "change_from_previous": null
      },
      {
        "period": "January 2025",
        "value": 625,
        "change_from_previous": 15.7
      },
      {
        "period": "February 2025",
        "value": 680,
        "change_from_previous": 8.8
      }
    ]
  }
  ```

---

## Learning Path

### Week 33: Advanced SQL Queries
- Window functions (ROW_NUMBER, RANK, LAG, LEAD)
- Common Table Expressions (CTEs)
- Complex aggregations
- Date/time functions

### Week 34: Goal Tracking System
- Goal types (distance, duration, frequency)
- Goal periods (weekly, monthly, yearly)
- Progress tracking
- Achievement notifications

### Week 35: PDF Report Generation
- Generate monthly/yearly reports
- Charts and visualizations
- User statistics summary
- Export to PDF

### Week 36: Chart Data Endpoints
- Time-series data for charts
- Activity type distribution
- Trend analysis
- Comparison data (vs previous period)

---

## Advanced SQL

### Monthly Statistics
```sql
-- Monthly summary with aggregates
SELECT
    DATE_TRUNC('month', activity_date) as month,
    COUNT(*) as activities,
    SUM(duration_minutes) as total_duration,
    SUM(distance_km) as total_distance,
    AVG(duration_minutes) as avg_duration,
    SUM(calories_burned) as total_calories
FROM activities
WHERE user_id = $1 AND deleted_at IS NULL
GROUP BY month
ORDER BY month DESC;
```

### Activity Type Distribution
```sql
-- Percentage breakdown by activity type
SELECT
    activity_type,
    COUNT(*) as count,
    SUM(duration_minutes) as total_duration,
    SUM(distance_km) as total_distance,
    ROUND(100.0 * COUNT(*) / SUM(COUNT(*)) OVER (), 2) as percentage
FROM activities
WHERE user_id = $1 AND deleted_at IS NULL
GROUP BY activity_type
ORDER BY count DESC;
```

### Streak Calculation
```sql
-- Calculate current activity streak
WITH daily_activities AS (
    SELECT DISTINCT DATE(activity_date) as activity_day
    FROM activities
    WHERE user_id = $1 AND deleted_at IS NULL
),
gaps AS (
    SELECT
        activity_day,
        activity_day - (ROW_NUMBER() OVER (ORDER BY activity_day))::integer AS gap_group
    FROM daily_activities
),
streaks AS (
    SELECT
        gap_group,
        COUNT(*) as streak_length,
        MIN(activity_day) as streak_start,
        MAX(activity_day) as streak_end
    FROM gaps
    GROUP BY gap_group
)
SELECT
    streak_length,
    streak_start,
    streak_end
FROM streaks
WHERE streak_end = CURRENT_DATE
ORDER BY streak_length DESC
LIMIT 1;
```

### Personal Records
```sql
-- Find personal records (longest distance, duration, etc.)
WITH ranked_activities AS (
    SELECT
        activity_type,
        activity_date,
        duration_minutes,
        distance_km,
        calories_burned,
        ROW_NUMBER() OVER (PARTITION BY activity_type ORDER BY distance_km DESC) as distance_rank,
        ROW_NUMBER() OVER (PARTITION BY activity_type ORDER BY duration_minutes DESC) as duration_rank,
        ROW_NUMBER() OVER (PARTITION BY activity_type ORDER BY calories_burned DESC) as calories_rank
    FROM activities
    WHERE user_id = $1 AND deleted_at IS NULL
)
SELECT
    activity_type,
    MAX(CASE WHEN distance_rank = 1 THEN distance_km END) as max_distance,
    MAX(CASE WHEN duration_rank = 1 THEN duration_minutes END) as max_duration,
    MAX(CASE WHEN calories_rank = 1 THEN calories_burned END) as max_calories
FROM ranked_activities
GROUP BY activity_type;
```

### Comparison with Previous Period
```sql
-- Compare current month with previous month
WITH current_month AS (
    SELECT
        COUNT(*) as activities,
        SUM(duration_minutes) as duration,
        SUM(distance_km) as distance
    FROM activities
    WHERE user_id = $1
        AND activity_date >= DATE_TRUNC('month', CURRENT_DATE)
        AND deleted_at IS NULL
),
previous_month AS (
    SELECT
        COUNT(*) as activities,
        SUM(duration_minutes) as duration,
        SUM(distance_km) as distance
    FROM activities
    WHERE user_id = $1
        AND activity_date >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
        AND activity_date < DATE_TRUNC('month', CURRENT_DATE)
        AND deleted_at IS NULL
)
SELECT
    c.activities as current_activities,
    p.activities as previous_activities,
    c.activities - p.activities as activities_change,
    ROUND(100.0 * (c.activities - p.activities) / NULLIF(p.activities, 0), 2) as activities_change_pct,
    c.duration as current_duration,
    p.duration as previous_duration,
    c.duration - p.duration as duration_change,
    c.distance as current_distance,
    p.distance as previous_distance,
    c.distance - p.distance as distance_change
FROM current_month c, previous_month p;
```

---

## Goal System

### Database Schema
```sql
CREATE TABLE goals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    goal_type VARCHAR(50) NOT NULL,  -- distance, duration, frequency, calories
    target_value DECIMAL NOT NULL,
    current_value DECIMAL DEFAULT 0,
    period VARCHAR(20) NOT NULL,     -- daily, weekly, monthly, yearly
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'active', -- active, completed, failed, abandoned
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_goals_user ON goals(user_id, status);
CREATE INDEX idx_goals_period ON goals(period, start_date, end_date);
```

### Goal Implementation
```go
type Goal struct {
    ID           int       `json:"id"`
    UserID       int       `json:"user_id"`
    GoalType     string    `json:"goal_type"`
    TargetValue  float64   `json:"target_value"`
    CurrentValue float64   `json:"current_value"`
    Period       string    `json:"period"`
    StartDate    time.Time `json:"start_date"`
    EndDate      time.Time `json:"end_date"`
    Status       string    `json:"status"`
    Progress     float64   `json:"progress"` // Calculated field
}

// Create a new goal
func (s *GoalService) CreateGoal(ctx context.Context, userID int, goalType string, targetValue float64, period string) (*Goal, error) {
    startDate, endDate := calculatePeriodDates(period)

    goal := &Goal{
        UserID:      userID,
        GoalType:    goalType,
        TargetValue: targetValue,
        Period:      period,
        StartDate:   startDate,
        EndDate:     endDate,
        Status:      "active",
    }

    if err := s.repo.Create(ctx, goal); err != nil {
        return nil, err
    }

    return goal, nil
}

// Update goal progress (called when activity is created)
func (s *GoalService) UpdateProgress(ctx context.Context, userID int, activityType string, value float64) error {
    // Get active goals for this user
    goals, err := s.repo.GetActiveGoals(ctx, userID)
    if err != nil {
        return err
    }

    for _, goal := range goals {
        // Check if this activity contributes to the goal
        if !goalAppliesToActivity(goal, activityType) {
            continue
        }

        // Update current value
        goal.CurrentValue += value

        // Calculate progress
        goal.Progress = (goal.CurrentValue / goal.TargetValue) * 100

        // Check if goal is completed
        if goal.CurrentValue >= goal.TargetValue {
            goal.Status = "completed"

            // Send achievement notification
            s.notificationService.SendGoalCompleted(ctx, userID, goal)
        }

        // Update in database
        if err := s.repo.Update(ctx, goal); err != nil {
            return err
        }
    }

    return nil
}

// Check for expired goals (run daily via cron)
func (s *GoalService) CheckExpiredGoals(ctx context.Context) error {
    goals, err := s.repo.GetActiveGoals(ctx, 0) // Get all active goals

    if err != nil {
        return err
    }

    now := time.Now()

    for _, goal := range goals {
        if now.After(goal.EndDate) {
            if goal.CurrentValue >= goal.TargetValue {
                goal.Status = "completed"
            } else {
                goal.Status = "failed"
            }

            if err := s.repo.Update(ctx, goal); err != nil {
                log.Printf("Error updating goal %d: %v", goal.ID, err)
            }
        }
    }

    return nil
}
```

---

## PDF Report Generation

```go
import "github.com/jung-kurt/gofpdf"

func (s *ReportService) GenerateMonthlyReport(ctx context.Context, userID int, month time.Time) ([]byte, error) {
    // Fetch data
    stats, err := s.statsRepo.GetMonthlyStats(ctx, userID, month)
    if err != nil {
        return nil, err
    }

    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    activities, err := s.activityRepo.GetByMonth(ctx, userID, month)
    if err != nil {
        return nil, err
    }

    // Create PDF
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    // Title
    pdf.SetFont("Arial", "B", 20)
    pdf.Cell(0, 10, "Monthly Activity Report")
    pdf.Ln(15)

    // User Info
    pdf.SetFont("Arial", "", 12)
    pdf.Cell(0, 8, fmt.Sprintf("User: %s", user.Name))
    pdf.Ln(6)
    pdf.Cell(0, 8, fmt.Sprintf("Period: %s", month.Format("January 2006")))
    pdf.Ln(12)

    // Summary Statistics Section
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(0, 10, "Summary Statistics")
    pdf.Ln(8)

    pdf.SetFont("Arial", "", 12)

    // Create table
    pdf.SetFillColor(240, 240, 240)
    pdf.CellFormat(90, 8, "Metric", "1", 0, "L", true, 0, "")
    pdf.CellFormat(90, 8, "Value", "1", 1, "L", true, 0, "")

    pdf.CellFormat(90, 8, "Total Activities", "1", 0, "L", false, 0, "")
    pdf.CellFormat(90, 8, fmt.Sprintf("%d", stats.TotalActivities), "1", 1, "L", false, 0, "")

    pdf.CellFormat(90, 8, "Total Distance", "1", 0, "L", false, 0, "")
    pdf.CellFormat(90, 8, fmt.Sprintf("%.2f km", stats.TotalDistance), "1", 1, "L", false, 0, "")

    pdf.CellFormat(90, 8, "Total Duration", "1", 0, "L", false, 0, "")
    pdf.CellFormat(90, 8, fmt.Sprintf("%d minutes", stats.TotalDuration), "1", 1, "L", false, 0, "")

    pdf.CellFormat(90, 8, "Calories Burned", "1", 0, "L", false, 0, "")
    pdf.CellFormat(90, 8, fmt.Sprintf("%d kcal", stats.TotalCalories), "1", 1, "L", false, 0, "")

    pdf.Ln(10)

    // Activity Breakdown by Type
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(0, 10, "Activity Breakdown")
    pdf.Ln(8)

    pdf.SetFont("Arial", "", 12)
    for activityType, count := range stats.ByType {
        pdf.Cell(0, 6, fmt.Sprintf("%s: %d activities", activityType, count))
        pdf.Ln(6)
    }

    pdf.Ln(10)

    // Personal Records
    if len(stats.PersonalRecords) > 0 {
        pdf.SetFont("Arial", "B", 16)
        pdf.Cell(0, 10, "Personal Records This Month")
        pdf.Ln(8)

        pdf.SetFont("Arial", "", 11)
        for _, pr := range stats.PersonalRecords {
            pdf.Cell(0, 6, fmt.Sprintf("- %s: %.2f km on %s", pr.Type, pr.Distance, pr.Date.Format("Jan 2")))
            pdf.Ln(6)
        }
    }

    // Convert to bytes
    var buf bytes.Buffer
    if err := pdf.Output(&buf); err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}
```

---

## Chart Data Endpoints

```go
// Time-series data for charts (last 30 days)
func (h *AnalyticsHandler) GetActivityTrend(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())
    days := 30 // Could be query parameter

    query := `
        SELECT
            DATE(activity_date) as date,
            COUNT(*) as count,
            SUM(duration_minutes) as duration,
            SUM(distance_km) as distance
        FROM activities
        WHERE user_id = $1
            AND activity_date >= CURRENT_DATE - INTERVAL '%d days'
            AND deleted_at IS NULL
        GROUP BY DATE(activity_date)
        ORDER BY date
    `

    rows, err := h.repo.db.Query(fmt.Sprintf(query, days), userID)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, "Query failed")
        return
    }
    defer rows.Close()

    type DataPoint struct {
        Date     string  `json:"date"`
        Count    int     `json:"count"`
        Duration int     `json:"duration"`
        Distance float64 `json:"distance"`
    }

    data := []DataPoint{}
    for rows.Next() {
        var dp DataPoint
        var date time.Time
        rows.Scan(&date, &dp.Count, &dp.Duration, &dp.Distance)
        dp.Date = date.Format("2006-01-02")
        data = append(data, dp)
    }

    response.JSON(w, http.StatusOK, data)
}

// Activity type distribution for pie chart
func (h *AnalyticsHandler) GetTypeDistribution(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r.Context())

    query := `
        SELECT
            activity_type,
            COUNT(*) as count
        FROM activities
        WHERE user_id = $1 AND deleted_at IS NULL
        GROUP BY activity_type
        ORDER BY count DESC
    `

    rows, err := h.repo.db.Query(query, userID)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, "Query failed")
        return
    }
    defer rows.Close()

    type TypeData struct {
        Type  string `json:"type"`
        Count int    `json:"count"`
    }

    data := []TypeData{}
    for rows.Next() {
        var td TypeData
        rows.Scan(&td.Type, &td.Count)
        data = append(data, td)
    }

    response.JSON(w, http.StatusOK, data)
}
```

---

## Common Pitfalls

1. **Not indexing date columns**
   - ❌ Slow time-based queries
   - ✅ Index `activity_date` column

2. **Calculating stats on every request**
   - ❌ Expensive queries repeated
   - ✅ Cache results, update incrementally

3. **Not handling timezones**
   - ❌ Incorrect date groupings
   - ✅ Store in UTC, convert for display

4. **Generating PDFs synchronously**
   - ❌ Slow API responses
   - ✅ Use background jobs

---

## Resources

- [PostgreSQL Window Functions](https://www.postgresql.org/docs/current/tutorial-window.html)
- [gofpdf Documentation](https://pkg.go.dev/github.com/jung-kurt/gofpdf)
- [Chart.js](https://www.chartjs.org/) (for frontend charts)

---

## Next Steps

After completing Month 9, you'll move to **Month 10: Containerization & CI/CD**, where you'll learn:
- Docker containerization
- Docker Compose for local development
- GitHub Actions for CI/CD
- Automated testing and deployment

**You now have powerful analytics!** 📊
