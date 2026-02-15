package utils

import (
	"io"
	"mime/multipart"
	"net/http"
)

func DetectFileType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Reset pointer so the file can be read again later (e.g., for S3 upload)
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	return http.DetectContentType(buffer), nil
}
