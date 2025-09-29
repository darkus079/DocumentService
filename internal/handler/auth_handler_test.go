package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"documentservice/internal/auth"
	"documentservice/internal/models"
	"documentservice/internal/repository"
	"documentservice/internal/service"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestHandler() (*DocumentHandler, func()) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	db := client.Database("test_documentservice")
	usersCollection := db.Collection("users")
	documentsCollection := db.Collection("documents")

	tokenMgr := auth.NewTokenManager(usersCollection, "admin123")
	docRepo := repository.NewDocumentRepository(documentsCollection)
	docService := service.NewDocumentService(docRepo, tokenMgr, 50*1024*1024)

	handler := NewDocumentHandler(docService, tokenMgr)

	cleanup := func() {
		usersCollection.Drop(context.Background())
		documentsCollection.Drop(context.Background())
		client.Disconnect(context.Background())
	}

	return handler, cleanup
}

func TestDocumentHandler_Register(t *testing.T) {
	handler, cleanup := setupTestHandler()
	defer cleanup()

	tests := []struct {
		name           string
		request        models.RegisterRequest
		expectedStatus int
	}{
		{
			name: "valid registration",
			request: models.RegisterRequest{
				Token: "admin123",
				Login: "testuser123",
				Pswd:  "Password123!",
			},
			expectedStatus: 200,
		},
		{
			name: "invalid admin token",
			request: models.RegisterRequest{
				Token: "wrong_token",
				Login: "testuser123",
				Pswd:  "Password123!",
			},
			expectedStatus: 400,
		},
		{
			name: "short login",
			request: models.RegisterRequest{
				Token: "admin123",
				Login: "short",
				Pswd:  "Password123!",
			},
			expectedStatus: 400,
		},
		{
			name: "weak password",
			request: models.RegisterRequest{
				Token: "admin123",
				Login: "testuser123",
				Pswd:  "weak",
			},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Register(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDocumentHandler_Auth(t *testing.T) {
	handler, cleanup := setupTestHandler()
	defer cleanup()

	_, err := handler.tokenMgr.RegisterUser("admin123", "testuser123", "Password123!")
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name           string
		request        models.AuthRequest
		expectedStatus int
	}{
		{
			name: "valid authentication",
			request: models.AuthRequest{
				Login: "testuser123",
				Pswd:  "Password123!",
			},
			expectedStatus: 200,
		},
		{
			name: "invalid credentials",
			request: models.AuthRequest{
				Login: "testuser123",
				Pswd:  "wrongpassword",
			},
			expectedStatus: 401,
		},
		{
			name: "non-existent user",
			request: models.AuthRequest{
				Login: "nonexistent",
				Pswd:  "Password123!",
			},
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.Auth(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDocumentHandler_Logout(t *testing.T) {
	handler, cleanup := setupTestHandler()
	defer cleanup()

	_, err := handler.tokenMgr.RegisterUser("admin123", "testuser123", "Password123!")
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	authResponse, err := handler.tokenMgr.AuthenticateUser("testuser123", "Password123!")
	if err != nil {
		t.Fatalf("Failed to authenticate test user: %v", err)
	}

	token := authResponse.Response.Token

	req := httptest.NewRequest("DELETE", "/api/auth/"+token, nil)
	w := httptest.NewRecorder()
	handler.Logout(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["response"].(map[string]interface{})[token] != true {
		t.Error("Expected token to be invalidated")
	}
}
