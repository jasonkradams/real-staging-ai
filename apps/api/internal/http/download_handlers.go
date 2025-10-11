package http

import (
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// presignImageDownloadHandler handles GET /api/v1/images/:id/presign
// Query params:
// - kind: original|staged (default: original)
// - expires_in: seconds (default: 600)
// - download: 1 to force Content-Disposition=attachment
func (s *Server) presignImageDownloadHandler(c echo.Context) error {
	imageID := c.Param("id")
	if imageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad_request", Message: "image id is required"})
	}

	kind := strings.ToLower(strings.TrimSpace(c.QueryParam("kind")))
	if kind == "" {
		kind = "original"
	}
	expiresIn := int64(600)
	if v := strings.TrimSpace(c.QueryParam("expires_in")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			expiresIn = n
		}
	}
	contentDisposition := ""
	if c.QueryParam("download") == "1" {
		contentDisposition = "attachment"
	}

	img, err := s.imageService.GetImageByID(c.Request().Context(), imageID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "image not found"})
	}

	var rawURL string
	if kind == "staged" {
		if img.StagedURL == nil || *img.StagedURL == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad_request", Message: "image has no staged_url"})
		}
		rawURL = *img.StagedURL
	} else {
		rawURL = img.OriginalURL
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Path == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad_request", Message: "invalid stored URL"})
	}

	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = os.Getenv("S3_BUCKET_NAME")
	}
	if bucket == "" {
		bucket = "real-staging"
	}

	// Derive file key from URL path. Expect path-style: /<bucket>/<key>
	p := strings.TrimPrefix(u.Path, "/")
	if !strings.HasPrefix(p, bucket+"/") {
		// fallback: if virtual-hosted style, path is already the key
	} else {
		p = strings.TrimPrefix(p, bucket+"/")
	}
	fileKey := p
	if fileKey == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad_request", Message: "could not derive file key"})
	}

	signed, err := s.s3Service.GeneratePresignedGetURL(c.Request().Context(), fileKey, expiresIn, contentDisposition)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_server_error", Message: "failed to presign URL"})
	}

	return c.JSON(http.StatusOK, map[string]string{"url": signed})
}
