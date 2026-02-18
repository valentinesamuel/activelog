package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/valentinesamuel/activelog/internal/platform/config"
	"github.com/valentinesamuel/activelog/internal/adapters/storage/types"
)

// ProviderName is the identifier for this storage provider
// const ProviderName = "s3"

// Provider implements the StorageProvider interface for AWS S3
type Provider struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

// New creates a new S3 storage provider
func New() (*Provider, error) {
	s3Cfg := config.Storage.S3

	if s3Cfg.Bucket == "" {
		return nil, fmt.Errorf("%w: S3 bucket not configured", types.ErrProviderNotConfigured)
	}

	// Build AWS config options
	var opts []func(*awsconfig.LoadOptions) error

	opts = append(opts, awsconfig.WithRegion(s3Cfg.Region))

	// Use explicit credentials if provided
	if s3Cfg.AccessKeyID != "" && s3Cfg.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				s3Cfg.AccessKeyID,
				s3Cfg.SecretAccessKey,
				"", // session token
			),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Build S3 client options
	var s3Opts []func(*s3.Options)

	// Custom endpoint for S3-compatible services (MinIO, LocalStack)
	if s3Cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s3Cfg.Endpoint)
		})
	}

	// Path-style for S3-compatible services
	if s3Cfg.UsePathStyle {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)
	presignClient := s3.NewPresignClient(client)

	return &Provider{
		client:        client,
		presignClient: presignClient,
		bucket:        s3Cfg.Bucket,
	}, nil
}

// Upload stores a file in S3 and returns metadata about the stored object
func (p *Provider) Upload(ctx context.Context, input *types.UploadInput) (*types.UploadOutput, error) {
	if input.Key == "" {
		return nil, types.ErrInvalidKey
	}

	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(input.Key),
		Body:        input.Body,
		ContentType: aws.String(input.ContentType),
	}

	// Add custom metadata if provided
	if len(input.Metadata) > 0 {
		putInput.Metadata = input.Metadata
	}

	result, err := p.client.PutObject(ctx, putInput)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", types.ErrUploadFailed, err)
	}

	etag := ""
	if result.ETag != nil {
		etag = *result.ETag
	}

	return &types.UploadOutput{
		Key:        input.Key,
		ETag:       etag,
		URL:        p.buildObjectURL(input.Key),
		UploadedAt: time.Now(),
	}, nil
}

// Download retrieves a file from S3
func (p *Provider) Download(ctx context.Context, key string) (io.ReadCloser, *types.FileMetadata, error) {
	if key == "" {
		return nil, nil, types.ErrInvalidKey
	}

	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil, types.ErrNotFound
		}
		if isAccessDeniedError(err) {
			return nil, nil, types.ErrAccessDenied
		}
		return nil, nil, fmt.Errorf("failed to download object: %w", err)
	}

	metadata := &types.FileMetadata{
		Key:         key,
		Size:        aws.ToInt64(result.ContentLength),
		ContentType: aws.ToString(result.ContentType),
		ETag:        aws.ToString(result.ETag),
	}
	if result.LastModified != nil {
		metadata.LastModified = *result.LastModified
	}

	return result.Body, metadata, nil
}

// Delete removes a file from S3
func (p *Provider) Delete(ctx context.Context, key string) error {
	if key == "" {
		return types.ErrInvalidKey
	}

	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isAccessDeniedError(err) {
			return types.ErrAccessDenied
		}
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// DeleteMultiple removes multiple files efficiently using batch delete
func (p *Provider) DeleteMultiple(ctx context.Context, keys []string) (map[string]error, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	// Build delete objects input
	objects := make([]s3types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = s3types.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	result, err := p.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(p.bucket),
		Delete: &s3types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete objects: %w", err)
	}

	// Collect errors for individual objects
	errors := make(map[string]error)
	for _, objErr := range result.Errors {
		key := aws.ToString(objErr.Key)
		msg := aws.ToString(objErr.Message)
		errors[key] = fmt.Errorf("failed to delete %s: %s", key, msg)
	}

	if len(errors) > 0 {
		return errors, nil
	}
	return nil, nil
}

// List returns files matching the given prefix
func (p *Provider) List(ctx context.Context, input *types.ListInput) (*types.ListOutput, error) {
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(p.bucket),
	}

	if input.Prefix != "" {
		listInput.Prefix = aws.String(input.Prefix)
	}
	if input.MaxKeys > 0 {
		listInput.MaxKeys = aws.Int32(int32(input.MaxKeys))
	}
	if input.Marker != "" {
		listInput.ContinuationToken = aws.String(input.Marker)
	}

	result, err := p.client.ListObjectsV2(ctx, listInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	files := make([]types.FileMetadata, len(result.Contents))
	for i, obj := range result.Contents {
		files[i] = types.FileMetadata{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			ETag:         aws.ToString(obj.ETag),
			LastModified: aws.ToTime(obj.LastModified),
		}
	}

	output := &types.ListOutput{
		Files:       files,
		IsTruncated: aws.ToBool(result.IsTruncated),
	}
	if result.NextContinuationToken != nil {
		output.NextMarker = *result.NextContinuationToken
	}

	return output, nil
}

// Exists checks if a file exists in S3
func (p *Provider) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, types.ErrInvalidKey
	}

	_, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// GetPresignedURL generates a time-limited URL for direct access
func (p *Provider) GetPresignedURL(ctx context.Context, input *types.PresignedURLInput) (string, error) {
	if input.Key == "" {
		return "", types.ErrInvalidKey
	}

	expiry := input.ExpiresIn
	if expiry == 0 {
		expiry = 15 * time.Minute // Default expiry
	}

	var url string
	var err error

	switch input.Operation {
	case types.PresignGet:
		result, presignErr := p.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(p.bucket),
			Key:    aws.String(input.Key),
		}, s3.WithPresignExpires(expiry))
		if presignErr != nil {
			err = presignErr
		} else {
			url = result.URL
		}

	case types.PresignPut:
		result, presignErr := p.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(p.bucket),
			Key:    aws.String(input.Key),
		}, s3.WithPresignExpires(expiry))
		if presignErr != nil {
			err = presignErr
		} else {
			url = result.URL
		}

	default:
		return "", fmt.Errorf("unsupported presign operation: %s", input.Operation)
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, nil
}

// GetMetadata retrieves file metadata without downloading the file
func (p *Provider) GetMetadata(ctx context.Context, key string) (*types.FileMetadata, error) {
	if key == "" {
		return nil, types.ErrInvalidKey
	}

	result, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil, types.ErrNotFound
		}
		if isAccessDeniedError(err) {
			return nil, types.ErrAccessDenied
		}
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	metadata := &types.FileMetadata{
		Key:         key,
		Size:        aws.ToInt64(result.ContentLength),
		ContentType: aws.ToString(result.ContentType),
		ETag:        aws.ToString(result.ETag),
	}
	if result.LastModified != nil {
		metadata.LastModified = *result.LastModified
	}

	return metadata, nil
}

// buildObjectURL constructs the public URL for an object
func (p *Provider) buildObjectURL(key string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", p.bucket, key)
}

// isNotFoundError checks if the error is a not found error
func isNotFoundError(err error) bool {
	var notFound *s3types.NotFound
	var noSuchKey *s3types.NoSuchKey
	return errorIs(err, &notFound) || errorIs(err, &noSuchKey)
}

// isAccessDeniedError checks if the error is an access denied error
func isAccessDeniedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "AccessDenied") || contains(errStr, "access denied")
}

// errorIs is a helper for checking AWS error types
func errorIs[T any](err error, target *T) bool {
	var t T
	return err != nil && (err == any(t) || contains(err.Error(), fmt.Sprintf("%T", t)))
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && containsImpl(s, substr)))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
