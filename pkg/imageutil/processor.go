package imageutil

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"

	"github.com/disintegration/imaging"
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

func ResizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	dstImageFit := imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)

	return dstImageFit
}

func ConvertToJPEG(img image.Image, format string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeImage(file multipart.File) (image.Image, error) {
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode image: %w", err)
	}

	return img, nil
}

func EncodeImage(img image.Image, format string) ([]byte, error) {
	switch format {
	case "jpg", "jpeg":
		bufferBytes, err := ConvertToJPEG(img, "jpeg")
		return bufferBytes, err
	case "png":
		buf := new(bytes.Buffer)
		err := png.Encode(buf, img)
		return buf.Bytes(), err
	default:
		return nil, fmt.Errorf("unsupported format")
	}

}

func GenerateThumbnail(img image.Image) image.Image {
	return ResizeImage(img, 300, 300)
}
