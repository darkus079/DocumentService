package service

import (
	"context"
	"fmt"
	"mime"
	"net/http"

	"documentservice/internal/auth"
	"documentservice/internal/models"
	"documentservice/internal/repository"
)

// DocumentService handles document business logic
type DocumentService struct {
	docRepo     *repository.DocumentRepository
	tokenMgr    *auth.TokenManager
	maxFileSize int64
}

// NewDocumentService creates a new document service
func NewDocumentService(docRepo *repository.DocumentRepository, tokenMgr *auth.TokenManager, maxFileSize int64) *DocumentService {
	return &DocumentService{
		docRepo:     docRepo,
		tokenMgr:    tokenMgr,
		maxFileSize: maxFileSize,
	}
}

// GetDocuments retrieves documents with authentication and filtering
func (ds *DocumentService) GetDocuments(ctx context.Context, params models.QueryParams) (*models.APIResponse, error) {
	// Validate token
	userLogin, err := ds.tokenMgr.ValidateToken(params.Token)
	if err != nil {
		return models.NewErrorResponse(401, "Unauthorized"), nil
	}

	// If no login specified, use the token owner's login
	if params.Login == "" {
		params.Login = userLogin
	}

	// Get documents
	documents, err := ds.docRepo.GetDocuments(ctx, params)
	if err != nil {
		return models.NewErrorResponse(500, "Internal server error"), err
	}

	// Filter documents based on access rights
	var accessibleDocs []models.Document
	for _, doc := range documents {
		hasAccess, err := ds.docRepo.CheckDocumentAccess(ctx, doc.ID, userLogin)
		if err != nil {
			continue // Skip documents with access check errors
		}
		if hasAccess {
			accessibleDocs = append(accessibleDocs, doc)
		}
	}

	response := models.DocumentList{Docs: accessibleDocs}
	return models.NewSuccessResponse(response), nil
}

// GetDocument retrieves a single document with authentication
func (ds *DocumentService) GetDocument(ctx context.Context, docID, token string) (*models.APIResponse, *http.Response, error) {
	// Validate token
	userLogin, err := ds.tokenMgr.ValidateToken(token)
	if err != nil {
		return models.NewErrorResponse(401, "Unauthorized"), nil, nil
	}

	// Check document access
	hasAccess, err := ds.docRepo.CheckDocumentAccess(ctx, docID, userLogin)
	if err != nil {
		return models.NewErrorResponse(500, "Internal server error"), nil, err
	}
	if !hasAccess {
		return models.NewErrorResponse(403, "Access denied"), nil, nil
	}

	// Get document
	document, err := ds.docRepo.GetDocumentByID(ctx, docID)
	if err != nil {
		if err.Error() == "document not found" {
			return models.NewErrorResponse(404, "Document not found"), nil, nil
		}
		return models.NewErrorResponse(500, "Internal server error"), nil, err
	}

	// If it's a file, return the file content
	if document.File && len(document.Content) > 0 {
		// Set appropriate MIME type
		mimeType := document.MIME
		if mimeType == "" {
			mimeType = mime.TypeByExtension("." + getFileExtension(document.Name))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
		}

		// Create HTTP response for file content
		httpResp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
		}
		httpResp.Header.Set("Content-Type", mimeType)
		httpResp.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", document.Name))

		return models.NewSuccessResponse(document), httpResp, nil
	}

	// Return JSON response for non-file documents
	return models.NewSuccessResponse(document), nil, nil
}

// ValidateFileSize validates if file size is within limits
func (ds *DocumentService) ValidateFileSize(size int64) error {
	if size > ds.maxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", ds.maxFileSize)
	}
	return nil
}

// ValidateMIMEType validates MIME type
func (ds *DocumentService) ValidateMIMEType(mimeType string) error {
	// Basic MIME type validation
	if mimeType == "" {
		return fmt.Errorf("MIME type is required")
	}

	// Check if it's a valid MIME type format
	_, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return fmt.Errorf("invalid MIME type format")
	}

	return nil
}

// getFileExtension extracts file extension from filename
func getFileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i+1:]
		}
	}
	return ""
}
