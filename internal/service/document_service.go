package service

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"time"

	"documentservice/internal/auth"
	"documentservice/internal/models"
	"documentservice/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocumentService struct {
	docRepo     *repository.DocumentRepository
	tokenMgr    *auth.TokenManager
	maxFileSize int64
}

func NewDocumentService(docRepo *repository.DocumentRepository, tokenMgr *auth.TokenManager, maxFileSize int64) *DocumentService {
	return &DocumentService{
		docRepo:     docRepo,
		tokenMgr:    tokenMgr,
		maxFileSize: maxFileSize,
	}
}

func (ds *DocumentService) GetDocuments(ctx context.Context, params models.QueryParams) (*models.APIResponse, error) {
	userLogin, err := ds.tokenMgr.ValidateToken(params.Token)
	if err != nil {
		return models.NewErrorResponse(401, "Unauthorized"), nil
	}

	if params.Login == "" {
		params.Login = userLogin
	}

	documents, err := ds.docRepo.GetDocuments(ctx, params)
	if err != nil {
		return models.NewErrorResponse(500, "Internal server error"), err
	}

	var accessibleDocs []models.Document
	for _, doc := range documents {
		hasAccess, err := ds.docRepo.CheckDocumentAccess(ctx, doc.ID, userLogin)
		if err != nil {
			continue
		}
		if hasAccess {
			accessibleDocs = append(accessibleDocs, doc)
		}
	}

	response := models.DocumentList{Docs: accessibleDocs}
	return models.NewSuccessResponse(response), nil
}

func (ds *DocumentService) GetDocument(ctx context.Context, docID, token string) (*models.APIResponse, *http.Response, error) {
	userLogin, err := ds.tokenMgr.ValidateToken(token)
	if err != nil {
		return models.NewErrorResponse(401, "Unauthorized"), nil, nil
	}

	hasAccess, err := ds.docRepo.CheckDocumentAccess(ctx, docID, userLogin)
	if err != nil {
		return models.NewErrorResponse(500, "Internal server error"), nil, err
	}
	if !hasAccess {
		return models.NewErrorResponse(403, "Access denied"), nil, nil
	}

	document, err := ds.docRepo.GetDocumentByID(ctx, docID)
	if err != nil {
		if err.Error() == "document not found" {
			return models.NewErrorResponse(404, "Document not found"), nil, nil
		}
		return models.NewErrorResponse(500, "Internal server error"), nil, err
	}

	if document.File && len(document.Content) > 0 {
		mimeType := document.MIME
		if mimeType == "" {
			mimeType = mime.TypeByExtension("." + getFileExtension(document.Name))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
		}

		httpResp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
		}
		httpResp.Header.Set("Content-Type", mimeType)
		httpResp.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", document.Name))

		return models.NewSuccessResponse(document), httpResp, nil
	}

	return models.NewSuccessResponse(document), nil, nil
}

func (ds *DocumentService) ValidateFileSize(size int64) error {
	if size > ds.maxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", ds.maxFileSize)
	}
	return nil
}

func (ds *DocumentService) ValidateMIMEType(mimeType string) error {
	if mimeType == "" {
		return fmt.Errorf("MIME type is required")
	}

	_, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return fmt.Errorf("invalid MIME type format")
	}

	return nil
}

func (ds *DocumentService) UploadDocument(ctx context.Context, meta models.DocumentMeta, jsonData interface{}, fileContent []byte) (*models.DocumentUploadResponse, error) {
	userLogin, err := ds.tokenMgr.ValidateToken(meta.Token)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	if err := ds.ValidateFileSize(int64(len(fileContent))); err != nil {
		return nil, err
	}

	if err := ds.ValidateMIMEType(meta.MIME); err != nil {
		return nil, err
	}

	docID := primitive.NewObjectID().Hex()
	document := &models.Document{
		ID:      docID,
		Name:    meta.Name,
		MIME:    meta.MIME,
		File:    meta.File,
		Public:  meta.Public,
		Grant:   meta.Grant,
		Owner:   userLogin,
		Content: fileContent,
		Created: time.Now(),
	}

	if err := ds.docRepo.CreateDocument(ctx, document); err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	response := &models.DocumentUploadResponse{}
	if jsonData != nil {
		response.Data.JSON = jsonData
	}
	if meta.File {
		response.Data.File = meta.Name
	}

	return response, nil
}

func (ds *DocumentService) DeleteDocument(ctx context.Context, docID, token string) (*models.DocumentDeleteResponse, error) {
	userLogin, err := ds.tokenMgr.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	hasAccess, err := ds.docRepo.CheckDocumentAccess(ctx, docID, userLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("access denied")
	}

	if err := ds.docRepo.DeleteDocument(ctx, docID); err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	response := &models.DocumentDeleteResponse{
		Response: map[string]bool{
			docID: true,
		},
	}

	return response, nil
}

func getFileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i+1:]
		}
	}
	return ""
}
