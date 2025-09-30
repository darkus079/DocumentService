package service

import (
	"context"
	"net/http"

	"documentservice/internal/models"
)

type DocumentServiceInterface interface {
	GetDocuments(ctx context.Context, params models.QueryParams) (*models.APIResponse, error)
	GetDocument(ctx context.Context, docID, token string) (*models.APIResponse, *http.Response, error)
	UploadDocument(ctx context.Context, meta models.DocumentMeta, jsonData interface{}, fileContent []byte) (*models.DocumentUploadResponse, error)
	DeleteDocument(ctx context.Context, docID, token string) (*models.DocumentDeleteResponse, error)
	ValidateFileSize(size int64) error
	ValidateMIMEType(mimeType string) error
}
