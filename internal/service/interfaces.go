package service

import (
	"context"
	"net/http"

	"documentservice/internal/models"
)

type DocumentServiceInterface interface {
	GetDocuments(ctx context.Context, params models.QueryParams) (*models.APIResponse, error)
	GetDocument(ctx context.Context, docID, token string) (*models.APIResponse, *http.Response, error)
	ValidateFileSize(size int64) error
	ValidateMIMEType(mimeType string) error
}
