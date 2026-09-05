package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/infrastructure/config"
	"awebo/app/pkg/idgen"
	"awebo/app/repositories"
)

type AvatarService struct {
	repo *repositories.UserRepository
	cfg  *config.Config
}

func NewAvatarService(repo *repositories.UserRepository, cfg *config.Config) *AvatarService {
	return &AvatarService{repo: repo, cfg: cfg}
}

// allowedAvatarTypes maps a sniffed MIME type to the file extension we save
// it under. Detected from content, not the client-supplied filename/header,
// so a mislabelled upload can't sneak past this check.
var allowedAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// Upload validates, stores, and immediately applies a new avatar image for
// userID, returning the updated session user. publicBaseURL is the
// scheme://host the request came in on, used to build an absolute URL.
func (s *AvatarService) Upload(userID uuid.UUID, file multipart.File, header *multipart.FileHeader, publicBaseURL string) (*entities.AuthUser, error) {
	maxBytes := s.cfg.MaxAvatarSizeMB * 1024 * 1024
	if header.Size > maxBytes {
		return nil, exceptions.BadRequest(fmt.Sprintf("Файл больше %d МБ.", s.cfg.MaxAvatarSizeMB))
	}

	sniff := make([]byte, 512)
	n, err := file.Read(sniff)
	if err != nil && err != io.EOF {
		return nil, exceptions.Internal("")
	}
	contentType := http.DetectContentType(sniff[:n])
	ext, ok := allowedAvatarTypes[contentType]
	if !ok {
		return nil, exceptions.BadRequest("Поддерживаются только изображения JPEG, PNG, WEBP или GIF.")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, exceptions.Internal("")
	}

	dir := filepath.Join(s.cfg.UploadsDir, "avatars")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, exceptions.Internal("")
	}

	filename := idgen.RandomHex(16) + ext
	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return nil, exceptions.Internal("")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, exceptions.Internal("")
	}

	url := strings.TrimRight(publicBaseURL, "/") + "/uploads/avatars/" + filename

	updated, err := s.repo.UpdateFields(userID, map[string]any{"avatar_path": url})
	if err != nil {
		return nil, exceptions.Internal("")
	}

	return toAuthUser(updated), nil
}
