package auth

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB() (*mongo.Collection, func()) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	db := client.Database("test_documentservice")
	collection := db.Collection("users")

	cleanup := func() {
		collection.Drop(context.Background())
		client.Disconnect(context.Background())
	}

	return collection, cleanup
}

func TestTokenManager_GenerateToken(t *testing.T) {
	collection, cleanup := setupTestDB()
	defer cleanup()

	tm := NewTokenManager(collection, "admin123")

	login := "testuser"
	token, err := tm.GenerateToken(login)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Generated token should not be empty")
	}

	user, err := tm.GetUserByToken(token)
	if err != nil {
		t.Fatalf("Failed to get user by token: %v", err)
	}

	if user.Login != login {
		t.Errorf("Expected login %s, got %s", login, user.Login)
	}
	if user.Token != token {
		t.Errorf("Expected token %s, got %s", token, user.Token)
	}
}

func TestTokenManager_ValidateToken(t *testing.T) {
	collection, cleanup := setupTestDB()
	defer cleanup()

	tm := NewTokenManager(collection, "admin123")

	login := "testuser"
	token, err := tm.GenerateToken(login)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	validatedLogin, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if validatedLogin != login {
		t.Errorf("Expected login %s, got %s", login, validatedLogin)
	}

	_, err = tm.ValidateToken("invalid_token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestTokenManager_GetUserByToken(t *testing.T) {
	collection, cleanup := setupTestDB()
	defer cleanup()

	tm := NewTokenManager(collection, "admin123")

	login := "testuser"
	token, err := tm.GenerateToken(login)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	user, err := tm.GetUserByToken(token)
	if err != nil {
		t.Fatalf("Failed to get user by token: %v", err)
	}

	if user.Login != login {
		t.Errorf("Expected login %s, got %s", login, user.Login)
	}

	_, err = tm.GetUserByToken("non_existing_token")
	if err == nil {
		t.Error("Expected error for non-existing token")
	}
}

func TestTokenManager_RegisterUser(t *testing.T) {
	collection, cleanup := setupTestDB()
	defer cleanup()

	tm := NewTokenManager(collection, "admin123")

	response, err := tm.RegisterUser("admin123", "newuser123", "Password123!")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	if response.Response.Login != "newuser123" {
		t.Errorf("Expected login %s, got %s", "newuser123", response.Response.Login)
	}

	_, err = tm.RegisterUser("wrong_token", "newuser456", "Password123!")
	if err == nil {
		t.Error("Expected error for invalid admin token")
	}

	_, err = tm.RegisterUser("admin123", "short", "Password123!")
	if err == nil {
		t.Error("Expected error for short login")
	}

	_, err = tm.RegisterUser("admin123", "validuser123", "weak")
	if err == nil {
		t.Error("Expected error for weak password")
	}
}

func TestTokenManager_AuthenticateUser(t *testing.T) {
	collection, cleanup := setupTestDB()
	defer cleanup()

	tm := NewTokenManager(collection, "admin123")

	_, err := tm.RegisterUser("admin123", "authuser123", "Password123!")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	response, err := tm.AuthenticateUser("authuser123", "Password123!")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	if response.Response.Token == "" {
		t.Error("Expected non-empty token")
	}

	_, err = tm.AuthenticateUser("authuser123", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password")
	}

	_, err = tm.AuthenticateUser("nonexistent", "Password123!")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestTokenManager_InvalidateToken(t *testing.T) {
	collection, cleanup := setupTestDB()
	defer cleanup()

	tm := NewTokenManager(collection, "admin123")

	_, err := tm.RegisterUser("admin123", "invalidateuser123", "Password123!")
	if err != nil {
		t.Fatalf("Failed to register user: %v", err)
	}

	authResponse, err := tm.AuthenticateUser("invalidateuser123", "Password123!")
	if err != nil {
		t.Fatalf("Failed to authenticate user: %v", err)
	}

	token := authResponse.Response.Token

	err = tm.InvalidateToken(token)
	if err != nil {
		t.Fatalf("Failed to invalidate token: %v", err)
	}

	_, err = tm.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for invalidated token")
	}
}
