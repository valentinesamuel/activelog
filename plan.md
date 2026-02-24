# Plan: Seed CSVs + Endpoint Test Checklist

## Context
The v3.0 relationship features have just been wired up. To verify every feature works end-to-end,
we need realistic seed data in every table and a structured call-order guide for all API endpoints
with copy-paste request bodies and query params.

---

## Part 1 — CSV Seed Files

**Location:** `scripts/seeds/` (new directory)

**Import order** (respects FK constraints):
1. `users.csv`
2. `tags.csv`
3. `activities.csv`
4. `activity_tags.csv`
5. `activity_photos.csv`
6. `comments.csv`
7. `webhooks.csv`
+ `import.sql` — runs all COPY commands in the correct order

System-managed tables (`daily_stats`, `exports`, `webhook_deliveries`, `inbox_event`, `outbox_event`)
are populated by the app automatically; no seed needed.

---

### `users.csv`
```
id,email,username,password_hash
1,alice@example.com,alice,$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi
2,bob@example.com,bob,$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi
3,carol@example.com,carol,$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi
```
> Password for all users: `password` (bcrypt hash above)

---

### `tags.csv`
```
id,name,parent_tag_id
1,exercise,
2,cardio,1
3,strength,1
4,morning-run,2
5,cycling,2
6,yoga,
```
> Tags 2-5 are children of `exercise` (id=1), demonstrating self-referential nesting.
> Querying `filter[tags.parent.name]=exercise` should return activities tagged with `cardio` or `strength`.

---

### `activities.csv`
```
id,user_id,activity_type,title,description,duration_minutes,distance_km,calories_burned,notes,activity_date
1,1,running,Morning Run,Early morning 5k run,30,5.00,300,Felt great today,2024-03-01T07:00:00Z
2,1,cycling,Evening Ride,Casual evening cycle,45,15.00,450,Nice weather,2024-03-02T18:00:00Z
3,2,yoga,Morning Yoga,Gentle yoga session,60,0.00,150,Very relaxing,2024-03-01T08:00:00Z
4,2,gym,Strength Training,Full body workout,75,0.00,500,Heavy lifts,2024-03-03T10:00:00Z
5,3,basketball,Basketball Practice,Team practice session,90,0.00,600,Good game,2024-03-02T16:00:00Z
6,3,walking,Evening Walk,Neighbourhood walk,40,3.50,200,Peaceful,2024-03-04T19:00:00Z
```

---

### `activity_tags.csv`
```
activity_id,tag_id
1,2
1,4
2,2
2,5
3,6
4,3
5,1
6,6
```
> Activity 1 (Morning Run) → cardio + morning-run
> Activity 2 (Evening Ride) → cardio + cycling
> Activity 3 (Yoga) → yoga
> Activity 4 (Strength Training) → strength
> Activity 5 (Basketball) → exercise
> Activity 6 (Evening Walk) → yoga

---

### `activity_photos.csv`
```
id,activity_id,s3_key,thumbnail_key,content_type,file_size,uploaded_at
1,1,photos/activity-1/photo.jpg,photos/activity-1/thumb.jpg,image/jpeg,102400,2024-03-01T07:30:00Z
2,2,photos/activity-2/photo.jpg,photos/activity-2/thumb.jpg,image/jpeg,204800,2024-03-02T18:45:00Z
3,3,photos/activity-3/photo.png,photos/activity-3/thumb.png,image/png,153600,2024-03-01T09:00:00Z
```
> `file_size` must be between 2 and 2457600 bytes (DB constraint).

---

### `comments.csv`
```
id,user_id,commentable_type,commentable_id,content
1,1,Activity,1,Great run this morning! Felt strong.
2,2,Activity,3,Yoga really helps with recovery.
3,1,Tag,2,Cardio is so important for heart health.
4,3,Tag,6,Yoga is my favourite tag!
```
> Rows 1-2 test polymorphic JOIN to `activities`.
> Rows 3-4 test polymorphic JOIN to `tags`.

---

### `webhooks.csv`
```
id,user_id,url,events,secret,active
550e8400-e29b-41d4-a716-446655440001,1,https://webhook.site/test-alice,"{activity.created,activity.deleted}",supersecret1,true
550e8400-e29b-41d4-a716-446655440002,2,https://webhook.site/test-bob,"{activity.created}",supersecret2,true
```
> `id` is UUID; `events` uses PostgreSQL array literal syntax.

---

### `import.sql`
```sql
-- Run from project root: psql $DATABASE_URL -f scripts/seeds/import.sql

BEGIN;

\COPY users(id, email, username, password_hash)
  FROM 'scripts/seeds/users.csv' CSV HEADER;

SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));

\COPY tags(id, name, parent_tag_id)
  FROM 'scripts/seeds/tags.csv' CSV HEADER NULL '';

SELECT setval('tags_id_seq', (SELECT MAX(id) FROM tags));

\COPY activities(id, user_id, activity_type, title, description,
  duration_minutes, distance_km, calories_burned, notes, activity_date)
  FROM 'scripts/seeds/activities.csv' CSV HEADER;

SELECT setval('activities_id_seq', (SELECT MAX(id) FROM activities));

\COPY activity_tags(activity_id, tag_id)
  FROM 'scripts/seeds/activity_tags.csv' CSV HEADER;

\COPY activity_photos(id, activity_id, s3_key, thumbnail_key,
  content_type, file_size, uploaded_at)
  FROM 'scripts/seeds/activity_photos.csv' CSV HEADER;

SELECT setval('activity_photos_id_seq', (SELECT MAX(id) FROM activity_photos));

\COPY comments(id, user_id, commentable_type, commentable_id, content)
  FROM 'scripts/seeds/comments.csv' CSV HEADER;

SELECT setval('comments_id_seq', (SELECT MAX(id) FROM comments));

\COPY webhooks(id, user_id, url, events, secret, active)
  FROM 'scripts/seeds/webhooks.csv' CSV HEADER;

COMMIT;
```

---

## Part 2 — Endpoint Test Reference

**Base URL:** `http://localhost:8080`
**Auth header (all protected routes):** `Authorization: Bearer <token>`

---

### Phase 0 — Setup

#### 1. Health Check
```
GET /health
```

---

#### 2. Register User
```
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "alice",
  "email": "alice@example.com",
  "password": "password"
}
```
> Skip if CSV already imported.

---

#### 3. Login — Get JWT ⭐ (run first, save token)
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "alice@example.com",
  "password": "password"
}
```
> Response: `{ "token": "...", "email": "alice@example.com" }`
> Save the token value for all subsequent requests.

---

### Phase 1 — Core Activity CRUD

#### 4. Create Activity
```
POST /api/v1/activities
Authorization: Bearer <token>
Content-Type: application/json

{
  "activityType": "running",
  "title": "Morning Run",
  "description": "Early morning 5k run",
  "durationMinutes": 30,
  "distanceKm": 5.0,
  "caloriesBurned": 300,
  "notes": "Felt great today",
  "activityDate": "2024-03-01T07:00:00Z"
}
```
> Save the returned `id` for subsequent calls.

---

#### 5. Get Single Activity
```
GET /api/v1/activities/1
Authorization: Bearer <token>
```

---

#### 6. Update Activity
```
PATCH /api/v1/activities/1
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Morning Run — Updated",
  "notes": "Even better than before",
  "caloriesBurned": 320
}
```
> All fields are optional; send only what you want to change.

---

#### 7. List Activities (no filters)
```
GET /api/v1/activities
Authorization: Bearer <token>
```

---

#### 8. Delete Activity
```
DELETE /api/v1/activities/1
Authorization: Bearer <token>
```
> Returns HTTP 204. Activity is soft-deleted (deleted_at is set).

---

### Phase 2 — Batch Operations

#### 9. Batch Create
```
POST /api/v1/activities/batch
Authorization: Bearer <token>
Content-Type: application/json

{
  "activities": [
    {
      "activityType": "cycling",
      "title": "Evening Ride",
      "description": "Casual evening cycle",
      "durationMinutes": 45,
      "distanceKm": 15.0,
      "caloriesBurned": 450,
      "notes": "Nice weather",
      "activityDate": "2024-03-02T18:00:00Z"
    },
    {
      "activityType": "yoga",
      "title": "Morning Yoga",
      "description": "Gentle yoga session",
      "durationMinutes": 60,
      "distanceKm": 0,
      "caloriesBurned": 150,
      "notes": "Very relaxing",
      "activityDate": "2024-03-03T08:00:00Z"
    },
    {
      "activityType": "gym",
      "title": "Strength Training",
      "description": "Full body workout",
      "durationMinutes": 75,
      "distanceKm": 0,
      "caloriesBurned": 500,
      "notes": "Heavy lifts",
      "activityDate": "2024-03-04T10:00:00Z"
    }
  ]
}
```
> Returns HTTP 207. Save the IDs from successful results.

---

#### 10. Batch Delete
```
DELETE /api/v1/activities/batch
Authorization: Bearer <token>
Content-Type: application/json

{
  "ids": [2, 3]
}
```
> Returns HTTP 207.

---

### Phase 3 — Filtering & v3.0 Features

#### 11. Basic column filter
```
GET /api/v1/activities?filter[activity_type]=running
Authorization: Bearer <token>
```

#### 12. ManyToMany JOIN — filter by tag name
```
GET /api/v1/activities?filter[tags.name]=cardio
Authorization: Bearer <token>
```

#### 13. WithConditions — soft-delete test
```sql
-- First, soft-delete the tag in psql:
UPDATE tags SET deleted_at = NOW() WHERE name = 'cardio';
```
```
GET /api/v1/activities?filter[tags.name]=cardio
Authorization: Bearer <token>
```
> Expected: 0 results — the WithConditions JOIN clause automatically excludes deleted tags.

#### 14. Cross-registry — filter by user username (Feature 2)
```
GET /api/v1/activities?filter[users.username]=alice
Authorization: Bearer <token>
```

#### 15. Cross-registry — filter by user email
```
GET /api/v1/activities?filter[users.email]=alice@example.com
Authorization: Bearer <token>
```

#### 16. Deep nesting self-referential — filter by parent tag (Feature 3)
```
GET /api/v1/activities?filter[tags.parent.name]=exercise
Authorization: Bearer <token>
```
> Expected: activities tagged with `cardio`, `strength`, `morning-run`, or `cycling`
> (all children of `exercise`). Triggers 3 JOINs: activities→activity_tags→tags→tags AS parent.

#### 17. Full-text search in title
```
GET /api/v1/activities?search[title]=morning
Authorization: Bearer <token>
```

#### 18. Cross-registry search by username
```
GET /api/v1/activities?search[users.username]=ali
Authorization: Bearer <token>
```

#### 19. Operator filter — date range
```
GET /api/v1/activities?filter[activity_date][gte]=2024-03-01T00:00:00Z&filter[activity_date][lte]=2024-03-03T23:59:59Z
Authorization: Bearer <token>
```

#### 20. Comparison filter + ordering
```
GET /api/v1/activities?filter[distance_km][gt]=5&order[distance_km]=DESC
Authorization: Bearer <token>
```

#### 21. Pagination
```
GET /api/v1/activities?page=1&limit=3
Authorization: Bearer <token>
```

---

### Phase 4 — Stats

#### 22. Activity-level stats (with date range)
```
GET /api/v1/activities/stats?startDate=2024-01-01T00:00:00Z&endDate=2024-12-31T23:59:59Z
Authorization: Bearer <token>
```

#### 23. Weekly stats
```
GET /api/v1/stats/weekly
Authorization: Bearer <token>
```

#### 24. Monthly stats
```
GET /api/v1/stats/monthly
Authorization: Bearer <token>
```

#### 25. Activity count by type
```
GET /api/v1/stats/by-type
Authorization: Bearer <token>
```

#### 26. User activity summary
```
GET /api/v1/users/me/summary
Authorization: Bearer <token>
```

#### 27. Top tags (with limit)
```
GET /api/v1/users/me/tags/top?limit=5
Authorization: Bearer <token>
```

---

### Phase 5 — Export

#### 28. Export activities as CSV
```
GET /api/v1/activities/export/csv
Authorization: Bearer <token>
```
> Response is a downloadable CSV file.

#### 29. Enqueue PDF export
```
POST /api/v1/activities/export/pdf
Authorization: Bearer <token>
```
> Response: `{ "job_id": "..." }` — save this `job_id`.

#### 30. Poll job status
```
GET /api/v1/jobs/<jobId>/status
Authorization: Bearer <token>
```
> Poll until `status` = `completed` or `failed`.

#### 31. Get download URL
```
GET /api/v1/jobs/<jobId>/download
Authorization: Bearer <token>
```
> Returns presigned S3 URL (valid 15 minutes).

---

### Phase 6 — Webhooks

#### 32. Create webhook
```
POST /api/v1/webhooks
Authorization: Bearer <token>
Content-Type: application/json

{
  "url": "https://webhook.site/your-unique-id",
  "events": ["activity.created", "activity.deleted"]
}
```
> Save the returned `id` and `secret`.

#### 33. List webhooks
```
GET /api/v1/webhooks
Authorization: Bearer <token>
```

#### 34. Delete webhook
```
DELETE /api/v1/webhooks/<id>
Authorization: Bearer <token>
```

---

### Phase 7 — Photos

#### 35. Upload activity photo
```
POST /api/v1/activities/1/photos
Authorization: Bearer <token>
Content-Type: multipart/form-data

photos: <YOUR_IMAGE_FILE>   ← attach a real image file here (JPEG, PNG, or WebP, max 50 MB total)
```
> Supports up to 5 files per request.

#### 36. Get activity photos
```
GET /api/v1/activities/1/photos
Authorization: Bearer <token>
```

---

### Phase 8 — Misc

#### 37. Feature flags
```
GET /api/v1/features
Authorization: Bearer <token>
```

#### 38. WebSocket connection
```
WS ws://localhost:8080/ws?token=<jwt>
```
> Connect with a WebSocket client (e.g. Postman, wscat).
> `wscat -c "ws://localhost:8080/ws?token=<jwt>"`

---

## Files to Create

| File                                | Description                                           |
| ----------------------------------- | ----------------------------------------------------- |
| `scripts/seeds/users.csv`           | 3 users (alice, bob, carol) — password: `password`    |
| `scripts/seeds/tags.csv`            | 6 tags with self-referential parent nesting           |
| `scripts/seeds/activities.csv`      | 6 activities across 3 users                           |
| `scripts/seeds/activity_tags.csv`   | 8 junction rows                                       |
| `scripts/seeds/activity_photos.csv` | 3 photo metadata rows (placeholder S3 keys)           |
| `scripts/seeds/comments.csv`        | 4 polymorphic comments (Activity + Tag)               |
| `scripts/seeds/webhooks.csv`        | 2 webhook subscriptions                               |
| `scripts/seeds/import.sql`          | `\COPY` commands to load all 7 CSVs + sequence resets |

---

## Verification Checks

| Check                                                     | Expected                                                      |
| --------------------------------------------------------- | ------------------------------------------------------------- |
| `filter[tags.parent.name]=exercise`                       | Returns activities tagged cardio/strength/morning-run/cycling |
| Soft-delete `cardio` tag, then `filter[tags.name]=cardio` | 0 results (WithConditions)                                    |
| `filter[users.username]=alice`                            | Only alice's activities                                       |
| `filter[distance_km][gt]=5`                               | Only cycling/running rows                                     |
| `page=2&limit=2`                                          | Second page of results                                        |
