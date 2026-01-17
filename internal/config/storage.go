package config

// StorageConfigType holds storage configuration
type StorageConfigType struct {
	Provider string
	S3       S3ConfigType
	// Add other providers as needed:
	// Azure    AzureConfig
	// Supabase SupabaseConfig
}

// S3ConfigType holds AWS S3 configuration
type S3ConfigType struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // For S3-compatible services like MinIO, LocalStack
	UsePathStyle    bool   // For S3-compatible services
}

// Storage is the global storage configuration instance
var Storage *StorageConfigType

// loadStorage loads storage configuration from environment variables
func loadStorage() *StorageConfigType {
	return &StorageConfigType{
		Provider: GetEnv("STORAGE_PROVIDER", ""),
		S3: S3ConfigType{
			Bucket:          GetEnv("AWS_S3_BUCKET", ""),
			Region:          GetEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:     GetEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: GetEnv("AWS_SECRET_ACCESS_KEY", ""),
			Endpoint:        GetEnv("AWS_S3_ENDPOINT", ""),
			UsePathStyle:    GetEnvBool("AWS_S3_PATH_STYLE", false),
		},
	}
}
