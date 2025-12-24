# MONTH 4: File Uploads & Cloud Storage

**Weeks:** 13-16
**Phase:** File Handling & Cloud Integration
**Theme:** Add media support and professional API documentation

---

## Overview

This month introduces file handling capabilities to your application. You'll learn how to process file uploads, integrate with AWS S3 for cloud storage, handle image processing, and document your API with OpenAPI/Swagger. By the end, users can attach photos to their activities and your API will have professional, interactive documentation.

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
