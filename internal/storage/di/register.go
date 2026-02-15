package di

import (
	"log"

	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/storage/s3"
	"github.com/valentinesamuel/activelog/internal/storage/types"
)

// RegisterStorage registers the storage provider in the DI container
// The provider is selected based on the STORAGE_PROVIDER configuration
func RegisterStorage(c *container.Container) {
	c.Register(StorageProviderKey, func(c *container.Container) (interface{}, error) {
		return createProvider(), nil
	})
}

// createProvider creates the appropriate storage provider based on configuration
func createProvider() types.StorageProvider {
	switch config.Storage.Provider {
	case "s3":
		provider, err := s3.New()
		if err != nil {
			log.Printf("Warning: Failed to initialize S3 provider: %v. Storage operations will fail.", err)
			return nil
		}
		log.Printf("💾 Storage provider initialized: S3 (bucket: %s)", config.Storage.S3.Bucket)
		return provider

	case "local":
		log.Printf("Warning: Local storage provider not yet implemented")
		return nil

	case "supabase":
		log.Printf("Warning: Supabase storage provider not yet implemented")
		return nil

	case "azure":
		log.Printf("Warning: Azure Blob storage provider not yet implemented")
		return nil

	default:
		log.Printf("Warning: Unknown storage provider '%s'. Storage operations will fail.", config.Storage.Provider)
		return nil
	}
}
