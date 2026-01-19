package imageutil

import (
	"image"
)

// **Task 1: Install Image Processing Library** (10 min)
// - [ ] Install imaging library: `go get github.com/disintegration/imaging`
// - [ ] Or alternative: `go get github.com/nfnt/resize`
// - [ ] Test import in a simple file

// **Task 2: Implement Image Resizing** (60 min)
// - [ ] Create `pkg/imageutil/processor.go`
// - [ ] Implement `ResizeImage(img image.Image, maxWidth, maxHeight) image.Image`
// - [ ] Use `imaging.Fit()` to maintain aspect ratio
// - [ ] Implement `GenerateThumbnail(img image.Image) image.Image` (300x300)
// - [ ] Handle different image formats (JPEG, PNG, WebP)
// - [ ] Add tests with sample images

// **Task 3: Implement Image Format Conversion** (30 min)
// - [ ] Add `ConvertToJPEG(img image.Image, quality int) ([]byte, error)`
// - [ ] Use `jpeg.Encode()` with quality setting
// - [ ] Implement `EncodeImage(img, format) ([]byte, error)` for flexibility
// - [ ] Test conversion maintains image quality

// **Task 4: Update Upload Handler with Image Processing** (90 min)
// - [ ] Modify `Upload` handler to decode uploaded images
// - [ ] Resize main image (max 1920x1080)
// - [ ] Generate thumbnail (300x300)
// - [ ] Upload both versions to S3:
//   - Main: `activities/{id}/{uuid}.jpg`
//   - Thumb: `activities/{id}/thumb_{uuid}.jpg`
// - [ ] Store both S3 keys in database
// - [ ] Implement cleanup on failure (delete both from S3)

func ResizeImage(img image.Image, maxWidth, maxHeight int64) image.Image {

	return nil
}

func ConvertToJPEG(img image.Image, quality int) ([]byte, error) {
	// dstImageFit := imaging.Fit(img, 800, 600, imaging.Lanczos)
	return nil, nil
}

func EncodeImage(img, format string) ([]byte, error) {
	return nil, nil

}

func GenerateThumbnail(img image.Image) image.Image {
	return nil
}
