package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"awebo/app/exceptions"
	"awebo/app/infrastructure/config"
	"awebo/app/pkg/idgen"
)

// allowedImageTypes maps a sniffed MIME type to the file extension we save
// it under. Detected from content, not the client-supplied filename/header,
// so a mislabelled upload can't sneak past this check.
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// UploadService is the shared image-storage primitive behind avatar uploads
// and article images (cover + inline content images) — same validation,
// same on-disk layout, just a different subdirectory per caller.
type UploadService struct {
	cfg *config.Config
}

func NewUploadService(cfg *config.Config) *UploadService {
	return &UploadService{cfg: cfg}
}

// SaveImage validates and stores an uploaded image under uploads/<subdir>/,
// returning a domain-relative URL (e.g. "/uploads/articles/xxx.jpg") — no
// scheme/host, so it survives domain/protocol changes; the frontend resolves
// it against whatever origin it's served from.
func (s *UploadService) SaveImage(subdir string, file multipart.File, header *multipart.FileHeader) (string, error) {
	maxBytes := s.cfg.MaxAvatarSizeMB * 1024 * 1024
	if header.Size > maxBytes {
		return "", exceptions.BadRequest(fmt.Sprintf("Файл больше %d МБ.", s.cfg.MaxAvatarSizeMB))
	}

	sniff := make([]byte, 512)
	n, err := file.Read(sniff)
	if err != nil && err != io.EOF {
		return "", exceptions.Internal("")
	}
	contentType := http.DetectContentType(sniff[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return "", exceptions.BadRequest("Поддерживаются только изображения JPEG, PNG, WEBP или GIF.")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", exceptions.Internal("")
	}

	dir := filepath.Join(s.cfg.UploadsDir, subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", exceptions.Internal("")
	}

	filename := idgen.RandomHex(16) + ext
	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return "", exceptions.Internal("")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", exceptions.Internal("")
	}

	return "/uploads/" + subdir + "/" + filename, nil
}
