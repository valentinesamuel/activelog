# MONTH 4: File Uploads & Cloud Storage

**Weeks:** 13-16
**Phase:** File Handling & Cloud Integration
**Theme:** Add media support and professional API documentation

---

## Overview

This month introduces file handling capabilities to your application. You'll learn how to process file uploads, integrate with AWS S3 for cloud storage, handle image processing, and document your API with OpenAPI/Swagger. By the end, users can attach photos to their activities and your API will have professional, interactive documentation.

---

## API Endpoints Reference (for Postman Testing)

### Photo Upload Endpoints

**Upload Photos to Activity:**
- **HTTP Method:** `POST`
- **URL:** `/api/v1/activities/{id}/photos`
- **Headers:**
  ```
  Content-Type: multipart/form-data
  Authorization: Bearer <your-jwt-token>
  ```
- **Request Body (multipart/form-data):**
  ```
  Form field: "photos" (file, multiple allowed - max 5)
  - photo1.jpg
  - photo2.png
  - photo3.webp
  ```
- **Postman Setup:**
  - Select "Body" tab
  - Choose "form-data"
  - Add key: `photos`, Type: `File`
  - Click "Select Files" and choose up to 5 images
  - Can add multiple `photos` fields or use multi-select
- **Success Response (201 Created):**
  ```json
  {
    "photos": [
      {
        "id": 45,
        "activity_id": 123,
        "s3_key": "activities/123/uuid-photo1.jpg",
        "thumbnail_key": "activities/123/thumb-uuid-photo1.jpg",
        "url": "https://bucket.s3.amazonaws.com/activities/123/uuid-photo1.jpg",
        "thumbnail_url": "https://bucket.s3.amazonaws.com/activities/123/thumb-uuid-photo1.jpg",
        "content_type": "image/jpeg",
        "file_size": 2457600,
        "uploaded_at": "2024-01-15T14:30:22Z"
      },
      {
        "id": 46,
        "activity_id": 123,
        "s3_key": "activities/123/uuid-photo2.png",
        "thumbnail_key": "activities/123/thumb-uuid-photo2.png",
        "url": "https://bucket.s3.amazonaws.com/activities/123/uuid-photo2.png",
        "thumbnail_url": "https://bucket.s3.amazonaws.com/activities/123/thumb-uuid-photo2.png",
        "content_type": "image/png",
        "file_size": 1856432,
        "uploaded_at": "2024-01-15T14:30:23Z"
      }
    ],
    "uploaded_count": 2
  }
  ```
- **Error Response (400 Bad Request - Too Many Files):**
  ```json
  {
    "error": "validation error",
    "message": "maximum 5 photos allowed per activity"
  }
  ```
- **Error Response (400 Bad Request - Invalid File Type):**
  ```json
  {
    "error": "validation error",
    "message": "invalid file type, only JPEG, PNG, and WebP images are allowed"
  }
  ```
- **Error Response (413 Payload Too Large):**
  ```json
  {
    "error": "file too large",
    "message": "file size exceeds maximum of 10MB"
  }
  ```

**Get Photos for Activity:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/activities/{id}/photos`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (200 OK):**
  ```json
  {
    "photos": [
      {
        "id": 45,
        "activity_id": 123,
        "url": "https://bucket.s3.amazonaws.com/activities/123/uuid-photo1.jpg",
        "thumbnail_url": "https://bucket.s3.amazonaws.com/activities/123/thumb-uuid-photo1.jpg",
        "content_type": "image/jpeg",
        "file_size": 2457600,
        "uploaded_at": "2024-01-15T14:30:22Z"
      },
      {
        "id": 46,
        "activity_id": 123,
        "url": "https://bucket.s3.amazonaws.com/activities/123/uuid-photo2.png",
        "thumbnail_url": "https://bucket.s3.amazonaws.com/activities/123/thumb-uuid-photo2.png",
        "content_type": "image/png",
        "file_size": 1856432,
        "uploaded_at": "2024-01-15T14:30:23Z"
      }
    ],
    "total": 2
  }
  ```

**Delete Photo:**
- **HTTP Method:** `DELETE`
- **URL:** `/api/v1/photos/{id}`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (204 No Content):**
  ```
  (empty body)
  ```
- **Error Response (404 Not Found):**
  ```json
  {
    "error": "not found",
    "message": "photo not found"
  }
  ```
- **Error Response (403 Forbidden):**
  ```json
  {
    "error": "forbidden",
    "message": "you can only delete photos from your own activities"
  }
  ```

### S3 Signed URL Endpoint (Week 14)

**Get Presigned Upload URL:**
- **HTTP Method:** `POST`
- **URL:** `/api/v1/activities/{id}/photos/upload-url`
- **Headers:**
  ```
  Content-Type: application/json
  Authorization: Bearer <your-jwt-token>
  ```
- **Request Body:**
  ```json
  {
    "filename": "my-photo.jpg",
    "content_type": "image/jpeg",
    "file_size": 2457600
  }
  ```
- **Success Response (200 OK):**
  ```json
  {
    "upload_url": "https://bucket.s3.amazonaws.com/activities/123/uuid-photo.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
    "s3_key": "activities/123/uuid-photo.jpg",
    "expires_at": "2024-01-15T15:30:22Z"
  }
  ```
  **Note:** Client uploads directly to this URL using PUT request, then confirms upload to your API

---

## Learning Path

### Week 13: Local File Upload Basics
- Multipart form data handling
- File validation (size, type)
- Temporary file storage
- File metadata extraction

### Week 14: AWS S3 Integration
- AWS SDK for Go setup
- S3 bucket configuration
- IAM permissions
- Upload files to S3

### Week 15: Image Processing + OpenAPI/Swagger Docs (60 min)
- Resize images before upload
- Generate thumbnails
- Image format conversion
- **NEW:** OpenAPI/Swagger API documentation

### Week 16: File Management + Cleanup
- List user's uploaded files
- Delete files from S3
- Cleanup orphaned files
- Storage quota management

---

# WEEKLY TASK BREAKDOWNS

## Week 13: Local File Upload Basics

### 📋 Implementation Tasks

**Task 1: Create Photo Model** (20 min)
- [X] Create `internal/models/photo.go`
- [X] Define Photo struct with fields: id, activity_id, s3_key, thumbnail_key, content_type, file_size, uploaded_at
- [X] Add JSON tags for API responses
- [X] Add validation tags for file size limits

**Task 2: Create Database Migration for Photos** (15 min)
- [X] Create migration `migrations/004_create_activity_photos.up.sql`
- [X] Add activity_photos table with foreign key to activities
- [X] Add ON DELETE CASCADE for cleanup
- [X] Create index on activity_id
- [X] Run migration

**Task 3: Implement Multipart Form Handler** (60 min)
- [X] Create `internal/handlers/photo_handler.go`
- [X] Implement `Upload(w, r)` method
  - **Logic:**
    1. Extract activity ID from URL path parameters
    2. Parse multipart form with `r.ParseMultipartForm(50 << 20)` - loads up to 50MB into memory
    3. Get files from `r.MultipartForm.File["photos"]` - returns slice of file headers
    4. Validate file count (reject if > 5 photos)
    5. For each file: open file, validate type/size, process (save temp or upload to S3)
    6. Return JSON response with uploaded file metadata (IDs, URLs, sizes)
- [X] Parse multipart form: `r.ParseMultipartForm(50 << 20)` (50MB limit)
- [X] Extract files from `r.MultipartForm.File["photos"]`
- [X] Validate file count (max 5 photos per activity)
- [X] Return file metadata in response

**Task 4: File Validation** (45 min)
- [X] Create `pkg/upload/validator.go`
- [X] Implement `ValidateFileType(contentType string) error`
  - **Logic:** Check if contentType is in allowed list (image/jpeg, image/png, image/webp). Return error if not. Also read first 512 bytes of file and use `http.DetectContentType()` to verify magic bytes match (prevents MIME type spoofing).
- [X] Check MIME type: accept image/jpeg, image/png, image/webp
- [X] Implement `ValidateFileSize(size int64) error` (max 10MB)
  - **Logic:** Check if size > 10*1024*1024 bytes. If yes, return error with message "file too large, maximum size is 10MB". If no, return nil.
- [X] Check magic bytes (not just extension) for security
- [X] Add tests for validation logic

**Task 5: Temporary File Storage** (30 min)
- [X] Create temp directory for uploads: `/tmp/activelog-uploads`
- [X] Save uploaded files temporarily
- [X] Generate unique filenames with UUID
- [X] Implement cleanup of temp files after processing
- [X] Handle concurrent uploads safely

**Task 6: Create Photo Repository** (45 min)
- [X] Create `internal/repository/photo_repository.go`
- [X] Implement `Create(ctx, photo) error`
  - **Logic:** INSERT photo into activity_photos table with activity_id, s3_key, thumbnail_key, content_type, file_size. Use RETURNING clause to get generated ID and timestamps. Update photo struct with returned values.
- [X] Implement `GetByActivityID(ctx, activityID) ([]*Photo, error)`
  - **Logic:** SELECT * FROM activity_photos WHERE activity_id = $1 ORDER BY uploaded_at DESC. Scan all rows into Photo slice. Return empty slice if no photos (not an error).
- [X] Implement `GetByID(ctx, id) (*Photo, error)`
  - **Logic:** SELECT * FROM activity_photos WHERE id = $1. Scan into Photo struct. Return sql.ErrNoRows if photo doesn't exist.
- [X] Implement `Delete(ctx, id) error`
  - **Logic:** DELETE FROM activity_photos WHERE id = $1. Check rows affected - if 0, return error "photo not found". This only deletes DB record, caller must delete S3 files separately.
- [ ] Write tests for repository methods

**Task 7: Wire Up Routes and Test** (30 min)
- [ ] Add route: `POST /api/v1/activities/:id/photos`
- [ ] Add auth middleware to protect upload endpoint
- [ ] Test upload with Postman/curl
- [ ] Verify files saved to temp directory
- [ ] Verify database records created

### 📦 Files You'll Create/Modify

```
migrations/
├── 004_create_activity_photos.up.sql   [CREATE]
└── 004_create_activity_photos.down.sql [CREATE]

internal/
├── models/
│   └── photo.go                   [CREATE]
├── repository/
│   ├── photo_repository.go        [CREATE]
│   └── photo_repository_test.go   [CREATE]
├── handlers/
│   ├── photo_handler.go           [CREATE]
│   └── photo_handler_test.go      [CREATE]

pkg/
└── upload/
    ├── validator.go               [CREATE]
    └── validator_test.go          [CREATE]
```

### 🔄 Implementation Order

1. **Database**: Migration → Photo model
2. **Validation**: File validator utility
3. **Repository**: Photo repository methods
4. **Handler**: Upload handler with multipart parsing
5. **Integration**: Wire routes → Manual testing

### ⚠️ Blockers to Watch For

- **Memory limits**: `ParseMultipartForm` loads into memory - use streaming for large files
- **MIME type spoofing**: Check magic bytes, not just Content-Type header
- **Temp file cleanup**: Always clean up temp files, even on error (use defer)
- **Concurrent uploads**: Use unique filenames (UUID) to avoid conflicts
- **File permissions**: Ensure temp directory has write permissions

### ✅ Definition of Done

- [ ] Can upload multiple photos (up to 5) per activity
- [ ] Only valid image types accepted (jpg, png, webp)
- [ ] Files larger than 10MB rejected
- [ ] Database records created for uploads
- [ ] Temp files cleaned up after processing
- [ ] All tests passing

---

## Week 14: AWS S3 Integration

### 📋 Implementation Tasks

**Task 1: Set Up AWS Account and S3 Bucket** (30 min)
- [X] Create/log into AWS account
- [X] Create S3 bucket: `activelog-uploads-[your-name]`
- [X] Set bucket to private (block all public access)
- [X] Create folder structure: `activities/{activity_id}/`
- [X] Note bucket name and region for config

**Task 2: Create IAM User and Permissions** (20 min)
- [X] Go to IAM console
- [X] Create new user: `activelog-api`
- [X] Create policy with S3 permissions (PutObject, GetObject, DeleteObject)
- [X] Attach policy to user
- [X] Generate access keys (Access Key ID + Secret)
- [X] **IMPORTANT**: Save keys securely, never commit to git

**Task 3: Configure AWS Credentials** (15 min)
- [X] Install AWS SDK: `go get github.com/aws/aws-sdk-go-v2/config`
- [X] Install S3 client: `go get github.com/aws/aws-sdk-go-v2/service/s3`
- [X] Add to `.env` file:
  ```
  AWS_ACCESS_KEY_ID=your_access_key
  AWS_SECRET_ACCESS_KEY=your_secret_key
  AWS_REGION=us-east-1
  AWS_S3_BUCKET=activelog-uploads-yourname
  ```
- [X] Add `.env` to `.gitignore` if not already

**Task 4: Implement S3 Client** (60 min)
- [X] Create `pkg/storage/s3_client.go`
- [X] Implement `NewS3Client(bucket, region) (*S3Client, error)`
  - **Logic:** Load AWS config using `config.LoadDefaultConfig(ctx)` which reads credentials from env vars or ~/.aws/credentials. Create S3 client from config with specified region. Return S3Client struct containing bucket name and S3 service client.
- [X] Load AWS credentials from environment
- [X] Implement `Upload(ctx, key, file, contentType) error`
  - **Logic:** Use `s3.PutObjectInput` with Bucket, Key, Body (file reader), ContentType. Call `s3Client.PutObject(ctx, input)`. Check for errors. Key is full path like "activities/123/uuid.jpg". File must be io.Reader. Return wrapped error on failure.
- [X] Implement `Delete(ctx, key) error`
  - **Logic:** Use `s3.DeleteObjectInput` with Bucket and Key. Call `s3Client.DeleteObject(ctx, input)`. Note: DeleteObject succeeds even if key doesn't exist (idempotent). Return wrapped error on failure.
- [X] Handle AWS errors and wrap with context

**Task 5: Implement Presigned URLs** (45 min)
- [X] Add `GetPresignedURL(ctx, key, duration) (string, error)` to S3Client
  - **Logic:**
    1. Create presign client with `s3.NewPresignClient(s3Client)`
    2. Build GetObjectInput with Bucket and Key
    3. Call `presignClient.PresignGetObject(ctx, input, func(opts *PresignOptions) { opts.Expires = duration })`
    4. Returns signed URL string that allows temporary public access to private S3 object
    5. URL expires after specified duration (typically 1 hour)
    - **Why:** S3 bucket is private, so direct links don't work. Presigned URLs grant temporary access without making bucket public.
- [X] Use `s3.NewPresignClient()` for presigning
- [X] Set expiration to 1 hour for view URLs
- [X] Test presigned URL generation
- [X] Verify URLs work in browser

**Task 6: Update Photo Handler to Use S3** (90 min)
- [X] Modify `Upload` handler to upload to S3 instead of temp storage
- [X] Generate S3 key: `activities/{activityID}/{uuid}.jpg`
- [X] Upload file to S3
- [X] Store S3 key in database (not local path)
- [X] Implement rollback: delete from S3 if database insert fails
- [X] Update temp file cleanup logic
**Task 7: Implement Download/View Endpoints** (45 min)
- [ ] Add `GET /api/v1/activities/:id/photos/:photoId` handler
- [ ] Fetch photo record from database
- [ ] Verify user owns the activity (authorization)
- [ ] Generate presigned URL for the S3 object
- [ ] Redirect to presigned URL or return URL in JSON
- [ ] Test in browser

### 📦 Files You'll Create/Modify

```
pkg/
└── storage/
    ├── s3_client.go               [CREATE]
    └── s3_client_test.go          [CREATE]

internal/
├── handlers/
│   └── photo_handler.go           [MODIFY - use S3]
└── config/
    └── config.go                  [MODIFY - add AWS config]

.env                               [MODIFY - add AWS creds]
.gitignore                         [MODIFY - ensure .env ignored]
```

### 🔄 Implementation Order

1. **AWS Setup**: Create bucket → IAM user → Get credentials
2. **Configuration**: Add env vars → Load in config
3. **S3 Client**: Implement upload/delete/presign methods
4. **Integration**: Update handler to use S3
5. **Testing**: Upload → Verify in S3 console → Test presigned URLs

### ⚠️ Blockers to Watch For

- **Credentials in code**: NEVER hardcode AWS keys - always use env vars
- **Bucket naming**: Must be globally unique, lowercase, no underscores
- **Permissions**: IAM policy must allow s3:PutObject, s3:GetObject, s3:DeleteObject
- **Region mismatch**: Client region must match bucket region
- **Presigned URL expiration**: URLs expire - don't store them, generate on-demand
- **Error handling**: AWS errors can be cryptic - log full error details

### ✅ Definition of Done

- [ ] S3 bucket created and configured
- [ ] IAM user with proper permissions
- [ ] Files upload successfully to S3
- [ ] Can generate presigned URLs that work
- [ ] S3 objects deleted if database insert fails
- [ ] No AWS credentials in git repository
- [ ] All tests passing with S3 integration

---

## Week 15: Image Processing + OpenAPI/Swagger Docs

### 📋 Implementation Tasks

**Task 1: Install Image Processing Library** (10 min)
- [ ] Install imaging library: `go get github.com/disintegration/imaging`
- [ ] Or alternative: `go get github.com/nfnt/resize`
- [ ] Test import in a simple file

**Task 2: Implement Image Resizing** (60 min)
- [ ] Create `pkg/imageutil/processor.go`
- [ ] Implement `ResizeImage(img image.Image, maxWidth, maxHeight) image.Image`
- [ ] Use `imaging.Fit()` to maintain aspect ratio
- [ ] Implement `GenerateThumbnail(img image.Image) image.Image` (300x300)
- [ ] Handle different image formats (JPEG, PNG, WebP)
- [ ] Add tests with sample images

**Task 3: Implement Image Format Conversion** (30 min)
- [ ] Add `ConvertToJPEG(img image.Image, quality int) ([]byte, error)`
- [ ] Use `jpeg.Encode()` with quality setting
- [ ] Implement `EncodeImage(img, format) ([]byte, error)` for flexibility
- [ ] Test conversion maintains image quality

**Task 4: Update Upload Handler with Image Processing** (90 min)
- [ ] Modify `Upload` handler to decode uploaded images
- [ ] Resize main image (max 1920x1080)
- [ ] Generate thumbnail (300x300)
- [ ] Upload both versions to S3:
  - Main: `activities/{id}/{uuid}.jpg`
  - Thumb: `activities/{id}/thumb_{uuid}.jpg`
- [ ] Store both S3 keys in database
- [ ] Implement cleanup on failure (delete both from S3)

**Task 5: Install Swagger Tools** (15 min)
- [ ] Install swag CLI: `go install github.com/swaggo/swag/cmd/swag@latest`
- [ ] Install http-swagger: `go get github.com/swaggo/http-swagger`
- [ ] Verify installation: `swag --version`

**Task 6: Add Swagger Annotations** (120 min)
- [ ] Add general API info in `cmd/api/main.go`:
  ```go
  // @title ActiveLog API
  // @version 1.0
  // @description Activity tracking API
  // @host localhost:8080
  // @BasePath /api/v1
  ```
- [ ] Document auth handlers (Register, Login)
- [ ] Document activity handlers (Create, List, Get, Update, Delete)
- [ ] Document photo handlers (Upload, List, Get, Delete)
- [ ] Add request/response examples
- [ ] Add authorization requirements

**Task 7: Generate and Serve Swagger UI** (30 min)
- [ ] Run `swag init -g cmd/api/main.go`
- [ ] Verify `docs/` folder created
- [ ] Add swagger route in main.go:
  ```go
  import httpSwagger "github.com/swaggo/http-swagger"
  router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
  ```
- [ ] Start server and visit `http://localhost:8080/swagger/index.html`
- [ ] Test API endpoints directly from Swagger UI

### 📦 Files You'll Create/Modify

```
pkg/
└── imageutil/
    ├── processor.go               [CREATE]
    └── processor_test.go          [CREATE]

internal/
├── handlers/
│   ├── photo_handler.go           [MODIFY - add image processing]
│   ├── auth_handler.go            [MODIFY - add swagger comments]
│   ├── activity_handler.go        [MODIFY - add swagger comments]
│   └── stats_handler.go           [MODIFY - add swagger comments]

cmd/api/
└── main.go                        [MODIFY - add swagger route & general info]

docs/                              [GENERATED]
├── docs.go
├── swagger.json
└── swagger.yaml
```

### 🔄 Implementation Order

1. **Image processing**: Install library → Implement resize/thumbnail → Tests
2. **Handler update**: Integrate processing into upload flow
3. **Swagger setup**: Install tools
4. **Documentation**: Add annotations to all handlers
5. **Generation**: Run swag init → Serve UI → Test

### ⚠️ Blockers to Watch For

- **Image memory**: Large images consume lots of memory - resize ASAP
- **Format support**: Not all formats supported equally (stick to JPG/PNG)
- **JPEG quality**: Balance file size vs quality (85-95 is good range)
- **Swagger regeneration**: Must run `swag init` after changing annotations
- **Swagger imports**: Generated docs must be imported in main.go
- **S3 cleanup**: If thumbnail upload fails, delete main image too

### ✅ Definition of Done

- [ ] Images resized to max 1920x1080 before S3 upload
- [ ] Thumbnails generated (300x300) for all photos
- [ ] Both versions stored in S3 with separate keys
- [ ] Swagger UI accessible at /swagger/
- [ ] All API endpoints documented in Swagger
- [ ] Can test endpoints directly from Swagger UI
- [ ] All tests passing

---

## Week 16: File Management + Cleanup

### 📋 Implementation Tasks

**Task 1: Implement List Photos Endpoint** (30 min)
- [ ] Add `ListPhotos(w, r)` handler
- [ ] Get activity_id from URL params
- [ ] Verify user owns the activity
- [ ] Fetch photos from repository
- [ ] Generate presigned URLs for each photo
- [ ] Return photos with URLs in JSON response

**Task 2: Implement Delete Photo Endpoint** (60 min)
- [ ] Add `DeletePhoto(w, r)` handler
- [ ] Get photo_id from URL params
- [ ] Fetch photo record from database
- [ ] Verify user owns the photo's activity (authorization)
- [ ] Delete main image from S3
- [ ] Delete thumbnail from S3
- [ ] Delete database record
- [ ] Handle errors (S3 delete failed, DB delete failed)
- [ ] Return 204 No Content on success

**Task 3: Implement Orphaned File Cleanup** (90 min)
- [ ] Create `internal/services/cleanup_service.go`
- [ ] Implement `FindOrphanedPhotos(ctx) ([]*Photo, error)`
  - Find photos in DB where activity doesn't exist
  - Or photos older than X days with no activity
- [ ] Implement `CleanupOrphanedFiles(ctx) error`
- [ ] Delete orphaned S3 objects
- [ ] Delete orphaned database records
- [ ] Log cleanup statistics

**Task 4: Implement Storage Quota System** (120 min)
- [ ] Add `storage_used` column to users table (migration)
- [ ] Create `internal/middleware/quota.go`
- [ ] Implement `CheckStorageQuota(next) http.Handler`
- [ ] Calculate user's current storage usage
- [ ] Compare against quota (e.g., 100MB for free tier)
- [ ] Reject uploads if over quota (402 Payment Required)
- [ ] Update storage_used when photos uploaded/deleted
- [ ] Add endpoint: `GET /api/v1/users/me/storage` to show usage

**Task 5: Implement Batch Delete** (45 min)
- [ ] Add `DELETE /api/v1/users/me/photos` endpoint
- [ ] Accept array of photo IDs in request body
- [ ] Verify user owns all photos
- [ ] Delete all photos in transaction
- [ ] Delete all S3 objects
- [ ] Update storage quota
- [ ] Return count of deleted photos

**Task 6: Add User Photos Endpoint** (30 min)
- [ ] Add `GET /api/v1/users/me/photos` endpoint
- [ ] List all photos for current user across all activities
- [ ] Support pagination (limit, offset)
- [ ] Generate presigned URLs
- [ ] Return with activity context (activity type, date)

**Task 7: Write Comprehensive Tests** (60 min)
- [ ] Test upload with quota enforcement
- [ ] Test delete removes both S3 objects
- [ ] Test orphaned file cleanup
- [ ] Test storage quota calculation
- [ ] Test authorization (can't delete other user's photos)
- [ ] Mock S3 client for unit tests

### 📦 Files You'll Create/Modify

```
migrations/
├── 005_add_storage_quota.up.sql   [CREATE]
└── 005_add_storage_quota.down.sql [CREATE]

internal/
├── services/
│   ├── cleanup_service.go         [CREATE]
│   └── cleanup_service_test.go    [CREATE]
├── middleware/
│   ├── quota.go                   [CREATE]
│   └── quota_test.go              [CREATE]
├── handlers/
│   └── photo_handler.go           [MODIFY - add list, delete, batch]
└── repository/
    └── photo_repository.go        [MODIFY - add orphan queries]

cmd/api/
└── main.go                        [MODIFY - add routes]
```

### 🔄 Implementation Order

1. **Basic CRUD**: List → Delete endpoints
2. **Cleanup**: Orphaned file detection → Cleanup service
3. **Quota**: Database migration → Middleware → Update on upload/delete
4. **Batch operations**: Batch delete endpoint
5. **Testing**: Comprehensive tests for all scenarios

### ⚠️ Blockers to Watch For

- **Authorization**: Always verify user owns resources before delete
- **S3 delete failures**: Might succeed in DB but fail in S3 - handle gracefully
- **Transaction scope**: Quota updates should be in transaction with photo operations
- **Orphan detection**: Be careful not to delete recently uploaded files
- **Presigned URL generation**: Don't generate for every photo in large lists (slow)
- **Cascade deletes**: When activity deleted, photos should auto-delete (ON DELETE CASCADE)

### ✅ Definition of Done

- [ ] Can list all photos for an activity
- [ ] Can delete individual photos (S3 + database)
- [ ] Can delete multiple photos at once
- [ ] Orphaned files cleaned up automatically
- [ ] Storage quota enforced on uploads
- [ ] Users can see their storage usage
- [ ] Can list all user's photos across activities
- [ ] All authorization checks working
- [ ] All tests passing

---

## AWS Services

- **S3 for storage**
  - Object storage for files
  - High durability (99.999999999%)
  - CDN-ready with CloudFront

- **IAM for permissions**
  - Least-privilege access
  - Service accounts for application
  - Temporary credentials with STS

- **Presigned URLs for security**
  - Time-limited access to private files
  - Direct upload from browser
  - No server bandwidth consumption

---

## Documentation

- 🔴 **OpenAPI/Swagger specification (swaggo/swag)**
  - Auto-generate API docs from code comments
  - Interactive API testing UI
  - Client SDK generation support

- **Auto-generated API documentation**
  - Always up-to-date with code
  - No manual documentation maintenance

- **Interactive API testing UI**
  - Test endpoints directly from browser
  - Share with frontend developers
  - Onboard new team members faster

---

## Implementation

### S3 Upload
```go
import (
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
    client *s3.Client
    bucket string
}

func NewS3Client(bucket string) (*S3Client, error) {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        return nil, err
    }

    return &S3Client{
        client: s3.NewFromConfig(cfg),
        bucket: bucket,
    }, nil
}

func (s *S3Client) Upload(ctx context.Context, key string, file io.Reader, contentType string) error {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        Body:        file,
        ContentType: aws.String(contentType),
    })
    return err
}
```

### Presigned URL Generation
```go
import "github.com/aws/aws-sdk-go-v2/service/s3"

func (s *S3Client) GetPresignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
    presignClient := s3.NewPresignClient(s.client)

    req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    }, func(opts *s3.PresignOptions) {
        opts.Expires = duration
    })

    if err != nil {
        return "", err
    }

    return req.URL, nil
}

// Usage: Generate 1-hour access link
url, err := s3Client.GetPresignedURL(ctx, "user/123/photo.jpg", time.Hour)
```

### Image Processing
```go
import (
    "image"
    "image/jpeg"
    "github.com/nfnt/resize"
)

func ResizeImage(img image.Image, maxWidth, maxHeight uint) image.Image {
    return resize.Thumbnail(maxWidth, maxHeight, img, resize.Lanczos3)
}

func GenerateThumbnail(img image.Image) image.Image {
    return ResizeImage(img, 300, 300)
}

// Example usage
func ProcessUploadedImage(file multipart.File) error {
    // Decode image
    img, format, err := image.Decode(file)
    if err != nil {
        return err
    }

    // Resize for display (max 1920x1080)
    resized := ResizeImage(img, 1920, 1080)

    // Generate thumbnail (300x300)
    thumbnail := GenerateThumbnail(img)

    // Upload both versions to S3
    // ... upload code ...

    return nil
}
```

### 🔴 OpenAPI/Swagger Documentation
```go
// Install: go get -u github.com/swaggo/swag/cmd/swag
// Install: go get -u github.com/swaggo/http-swagger

// Add to main.go
import httpSwagger "github.com/swaggo/http-swagger"

// Swagger endpoint
router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

// Document your handlers with comments:

// CreateActivity godoc
// @Summary Create a new activity
// @Description Create a new activity with optional photo uploads
// @Tags activities
// @Accept json,multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param body body CreateActivityRequest true "Activity data"
// @Param photos formData file false "Activity photos (max 5)"
// @Success 201 {object} ActivityResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/activities [post]
func (h *ActivityHandler) Create(w http.ResponseWriter, r *http.Request) {
    // ... implementation ...
}

// Generate docs: swag init -g cmd/api/main.go
// Access at: http://localhost:8080/swagger/index.html
```

---

## Features

### Activity Photos
```sql
CREATE TABLE activity_photos (
    id SERIAL PRIMARY KEY,
    activity_id INTEGER REFERENCES activities(id) ON DELETE CASCADE,
    s3_key VARCHAR(500) NOT NULL,
    thumbnail_key VARCHAR(500),
    content_type VARCHAR(50),
    file_size INTEGER,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_activity_photos_activity ON activity_photos(activity_id);
```

### API Endpoints
```
POST   /api/v1/activities/:id/photos     # Upload photos
GET    /api/v1/activities/:id/photos     # List photos
DELETE /api/v1/activities/:id/photos/:photoId  # Delete photo
GET    /api/v1/users/me/photos           # All user photos
```

### Features Checklist
- [x] Activity photos (multiple per activity - max 5)
- [x] Profile pictures
- [x] Thumbnails generation (300x300)
- [x] Automatic cleanup of orphaned files
- [x] Storage quota per user (enforce in middleware)
- [x] Supported formats: JPG, PNG, WebP
- [x] Max file size: 10MB per photo

---

## File Upload Handler Example
```go
func (h *PhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
    // Parse multipart form (max 50MB)
    if err := r.ParseMultipartForm(50 << 20); err != nil {
        response.Error(w, http.StatusBadRequest, "File too large")
        return
    }

    activityID := getActivityID(r)
    userID := getUserID(r.Context())

    // Get uploaded files
    files := r.MultipartForm.File["photos"]
    if len(files) > 5 {
        response.Error(w, http.StatusBadRequest, "Maximum 5 photos allowed")
        return
    }

    uploaded := []Photo{}

    for _, fileHeader := range files {
        // Validate file type
        if !isValidImageType(fileHeader.Header.Get("Content-Type")) {
            response.Error(w, http.StatusBadRequest, "Invalid file type")
            return
        }

        // Open file
        file, err := fileHeader.Open()
        if err != nil {
            response.Error(w, http.StatusInternalServerError, "Failed to read file")
            return
        }
        defer file.Close()

        // Decode image
        img, format, err := image.Decode(file)
        if err != nil {
            response.Error(w, http.StatusBadRequest, "Invalid image file")
            return
        }

        // Process images
        resized := ResizeImage(img, 1920, 1080)
        thumbnail := GenerateThumbnail(img)

        // Generate S3 keys
        key := fmt.Sprintf("activities/%d/%s.jpg", activityID, uuid.New())
        thumbKey := fmt.Sprintf("activities/%d/thumb_%s.jpg", activityID, uuid.New())

        // Upload to S3
        if err := h.s3.Upload(r.Context(), key, encodeJPEG(resized), "image/jpeg"); err != nil {
            response.Error(w, http.StatusInternalServerError, "Upload failed")
            return
        }

        if err := h.s3.Upload(r.Context(), thumbKey, encodeJPEG(thumbnail), "image/jpeg"); err != nil {
            // Cleanup main image if thumbnail fails
            h.s3.Delete(r.Context(), key)
            response.Error(w, http.StatusInternalServerError, "Upload failed")
            return
        }

        // Save to database
        photo := Photo{
            ActivityID:   activityID,
            S3Key:        key,
            ThumbnailKey: thumbKey,
            ContentType:  "image/jpeg",
            FileSize:     int(fileHeader.Size),
        }

        if err := h.repo.CreatePhoto(r.Context(), &photo); err != nil {
            // Cleanup S3 if DB fails
            h.s3.Delete(r.Context(), key)
            h.s3.Delete(r.Context(), thumbKey)
            response.Error(w, http.StatusInternalServerError, "Database error")
            return
        }

        uploaded = append(uploaded, photo)
    }

    response.JSON(w, http.StatusCreated, map[string]interface{}{
        "photos": uploaded,
    })
}
```

---

## Storage Quota Management
```go
// Middleware to check storage quota
func (m *QuotaMiddleware) CheckStorageQuota(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := getUserID(r.Context())

        // Get user's current storage usage
        usage, err := m.repo.GetStorageUsage(r.Context(), userID)
        if err != nil {
            http.Error(w, "Internal error", http.StatusInternalServerError)
            return
        }

        // Check against quota (e.g., 100MB for free tier)
        quota := m.getQuotaForUser(userID) // 100 * 1024 * 1024 for free users

        if usage >= quota {
            response.Error(w, http.StatusPaymentRequired, "Storage quota exceeded")
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

## Common Pitfalls

1. **Not validating file types**
   - ❌ Accepting any file extension
   - ✅ Check MIME type and magic bytes

2. **Storing files on local disk**
   - ❌ Running out of disk space
   - ✅ Use cloud storage (S3)

3. **No cleanup of failed uploads**
   - ❌ Orphaned files in S3
   - ✅ Cleanup S3 if database insert fails

4. **Exposing S3 URLs directly**
   - ❌ URLs expire or become public
   - ✅ Use presigned URLs with expiration

5. **Not compressing images**
   - ❌ Large files, slow loading
   - ✅ Resize and optimize before upload

---

## AWS Setup Checklist

- [ ] Create S3 bucket with private access
- [ ] Configure CORS for bucket (if uploading from browser)
- [ ] Create IAM user with S3-only permissions
- [ ] Generate access keys
- [ ] Store keys in environment variables (never in code!)
- [ ] Enable versioning on bucket (optional)
- [ ] Set up lifecycle rules for old files (optional)

---

## Testing

```go
func TestPhotoUpload(t *testing.T) {
    // Use LocalStack or MinIO for S3 testing
    s3Client := setupTestS3(t)
    defer s3Client.Cleanup()

    // Create test image
    img := createTestImage(t, 1000, 1000)
    body := encodeMultipartForm(t, "photo.jpg", img)

    req := httptest.NewRequest("POST", "/api/v1/activities/1/photos", body)
    req.Header.Set("Content-Type", "multipart/form-data; boundary=...")

    w := httptest.NewRecorder()
    handler.Upload(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
}
```

---

## Resources

- [AWS SDK for Go V2](https://aws.github.io/aws-sdk-go-v2/docs/)
- [S3 Best Practices](https://docs.aws.amazon.com/AmazonS3/latest/userguide/best-practices.html)
- [swaggo/swag Documentation](https://github.com/swaggo/swag)
- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [Go image package](https://pkg.go.dev/image)

---

## Next Steps

After completing Month 4, you'll move to **Month 5: Caching & Performance**, where you'll learn:
- Redis setup and caching
- Cache invalidation strategies
- Rate limiting
- Performance monitoring
- Soft deletes pattern

**You now have a feature-rich API with media support!** 📸
