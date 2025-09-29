package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"documentservice/internal/models"
	"documentservice/internal/service"
)

type MockDocumentService struct {
	GetDocumentsFunc   func(ctx context.Context, params models.QueryParams) (*models.APIResponse, error)
	GetDocumentFunc    func(ctx context.Context, docID, token string) (*models.APIResponse, *http.Response, error)
	UploadDocumentFunc func(ctx context.Context, meta models.DocumentMeta, jsonData interface{}, fileContent []byte) (*models.DocumentUploadResponse, error)
	DeleteDocumentFunc func(ctx context.Context, docID, token string) (*models.DocumentDeleteResponse, error)
}

func (m *MockDocumentService) GetDocuments(ctx context.Context, params models.QueryParams) (*models.APIResponse, error) {
	if m.GetDocumentsFunc != nil {
		return m.GetDocumentsFunc(ctx, params)
	}
	return models.NewErrorResponse(501, "Not implemented"), nil
}

func (m *MockDocumentService) GetDocument(ctx context.Context, docID, token string) (*models.APIResponse, *http.Response, error) {
	if m.GetDocumentFunc != nil {
		return m.GetDocumentFunc(ctx, docID, token)
	}
	return models.NewErrorResponse(501, "Not implemented"), nil, nil
}

func (m *MockDocumentService) ValidateFileSize(size int64) error {
	return nil
}

func (m *MockDocumentService) UploadDocument(ctx context.Context, meta models.DocumentMeta, jsonData interface{}, fileContent []byte) (*models.DocumentUploadResponse, error) {
	if m.UploadDocumentFunc != nil {
		return m.UploadDocumentFunc(ctx, meta, jsonData, fileContent)
	}
	return &models.DocumentUploadResponse{}, nil
}

func (m *MockDocumentService) DeleteDocument(ctx context.Context, docID, token string) (*models.DocumentDeleteResponse, error) {
	if m.DeleteDocumentFunc != nil {
		return m.DeleteDocumentFunc(ctx, docID, token)
	}
	return &models.DocumentDeleteResponse{}, nil
}

func (m *MockDocumentService) ValidateMIMEType(mimeType string) error {
	return nil
}

var _ service.DocumentServiceInterface = (*MockDocumentService)(nil)

func TestDocumentHandler_GetDocuments(t *testing.T) {
	mockService := &MockDocumentService{
		GetDocumentsFunc: func(ctx context.Context, params models.QueryParams) (*models.APIResponse, error) {
			if params.Token == "valid_token" {
				return models.NewSuccessResponse(models.DocumentList{Docs: []models.Document{}}), nil
			}
			return models.NewErrorResponse(401, "Unauthorized"), nil
		},
	}

	handler := NewDocumentHandler(mockService, nil)

	req := httptest.NewRequest("GET", "/api/docs?token=valid_token&limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetDocuments(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != nil {
		t.Errorf("Expected success response, got error: %v", response.Error)
	}

	req = httptest.NewRequest("GET", "/api/docs", nil)
	w = httptest.NewRecorder()

	handler.GetDocuments(w, req)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	mockService.GetDocumentsFunc = func(ctx context.Context, params models.QueryParams) (*models.APIResponse, error) {
		return models.NewErrorResponse(401, "Unauthorized"), nil
	}

	req = httptest.NewRequest("GET", "/api/docs?token=invalid_token", nil)
	w = httptest.NewRecorder()

	handler.GetDocuments(w, req)

	if w.Code != 401 {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestDocumentHandler_GetDocument(t *testing.T) {
	mockService := &MockDocumentService{
		GetDocumentFunc: func(ctx context.Context, docID, token string) (*models.APIResponse, *http.Response, error) {
			if token == "valid_token" && docID == "valid_id" {
				doc := models.Document{
					ID:   docID,
					Name: "test.txt",
					MIME: "text/plain",
				}
				return models.NewSuccessResponse(doc), nil, nil
			}
			if token == "valid_token" {
				return models.NewErrorResponse(404, "Document not found"), nil, nil
			}
			return models.NewErrorResponse(401, "Unauthorized"), nil, nil
		},
	}

	handler := NewDocumentHandler(mockService, nil)

	req := httptest.NewRequest("GET", "/api/docs/valid_id?token=valid_token", nil)
	w := httptest.NewRecorder()

	handler.GetDocument(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Error != nil {
		t.Errorf("Expected success response, got error: %v", response.Error)
	}

	req = httptest.NewRequest("GET", "/api/docs/valid_id", nil)
	w = httptest.NewRecorder()

	handler.GetDocument(w, req)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	req = httptest.NewRequest("GET", "/api/docs/?token=valid_token", nil)
	w = httptest.NewRecorder()

	handler.GetDocument(w, req)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	mockService.GetDocumentFunc = func(ctx context.Context, docID, token string) (*models.APIResponse, *http.Response, error) {
		return models.NewErrorResponse(401, "Unauthorized"), nil, nil
	}

	req = httptest.NewRequest("GET", "/api/docs/valid_id?token=invalid_token", nil)
	w = httptest.NewRecorder()

	handler.GetDocument(w, req)

	if w.Code != 401 {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestDocumentHandler_HealthCheck(t *testing.T) {
	handler := NewDocumentHandler(&MockDocumentService{}, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Response == nil {
		t.Error("Expected response message")
	}
}
