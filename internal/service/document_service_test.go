package service

import (
	"context"
	"testing"

	"documentservice/internal/auth"
	"documentservice/internal/models"
	"documentservice/internal/repository"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestService() (*DocumentService, func()) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	db := client.Database("test_documentservice")
	usersCollection := db.Collection("users")
	documentsCollection := db.Collection("documents")

	cleanup := func() {
		usersCollection.Drop(context.Background())
		documentsCollection.Drop(context.Background())
		client.Disconnect(context.Background())
	}

	tokenMgr := auth.NewTokenManager(usersCollection, "admin123")
	docRepo := repository.NewDocumentRepository(documentsCollection)
	docService := NewDocumentService(docRepo, tokenMgr, 50*1024*1024)

	return docService, cleanup
}

func TestDocumentService_ValidateFileSize(t *testing.T) {
	_, cleanup := setupTestService()
	defer cleanup()

	service := NewDocumentService(nil, nil, 50*1024*1024)

	err := service.ValidateFileSize(30 * 1024 * 1024)
	if err != nil {
		t.Errorf("Expected no error for valid file size, got: %v", err)
	}

	err = service.ValidateFileSize(60 * 1024 * 1024)
	if err == nil {
		t.Error("Expected error for invalid file size")
	}
}

func TestDocumentService_ValidateMIMEType(t *testing.T) {
	_, cleanup := setupTestService()
	defer cleanup()

	service := NewDocumentService(nil, nil, 50*1024*1024)

	validMIMEs := []string{
		"text/plain",
		"image/jpeg",
		"application/pdf",
		"application/json",
	}

	for _, mime := range validMIMEs {
		err := service.ValidateMIMEType(mime)
		if err != nil {
			t.Errorf("Expected no error for MIME type %s, got: %v", mime, err)
		}
	}

	invalidMIMEs := []string{
		"",
		"invalid",
		"text/",
		"/plain",
	}

	for _, mime := range invalidMIMEs {
		err := service.ValidateMIMEType(mime)
		if err == nil {
			t.Errorf("Expected error for MIME type %s", mime)
		}
	}
}

func TestDocumentService_GetDocuments(t *testing.T) {
	service, cleanup := setupTestService()
	defer cleanup()

	ctx := context.Background()
	token, err := service.tokenMgr.GenerateToken("testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	params := models.QueryParams{
		Token: token,
		Limit: 10,
	}

	response, err := service.GetDocuments(ctx, params)
	if err != nil {
		t.Fatalf("Failed to get documents: %v", err)
	}

	if response.Error != nil {
		t.Errorf("Expected success response, got error: %v", response.Error)
	}

	params.Token = "invalid_token"
	response, err = service.GetDocuments(ctx, params)
	if err != nil {
		t.Fatalf("Failed to get documents: %v", err)
	}

	if response.Error == nil || response.Error.Code != 401 {
		t.Error("Expected 401 error for invalid token")
	}
}

func TestDocumentService_GetDocument(t *testing.T) {
	service, cleanup := setupTestService()
	defer cleanup()

	ctx := context.Background()
	token, err := service.tokenMgr.GenerateToken("testuser")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	response, _, err := service.GetDocument(ctx, "nonexistent", token)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}

	if response.Error == nil || response.Error.Code != 404 {
		t.Error("Expected 404 error for non-existing document")
	}

	response, _, err = service.GetDocument(ctx, "someid", "invalid_token")
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}

	if response.Error == nil || response.Error.Code != 401 {
		t.Error("Expected 401 error for invalid token")
	}
}
