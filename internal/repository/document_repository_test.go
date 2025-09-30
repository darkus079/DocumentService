package repository

import (
	"context"
	"testing"

	"documentservice/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDocumentDB() (*mongo.Collection, func()) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	db := client.Database("test_documentservice")
	collection := db.Collection("documents")

	cleanup := func() {
		collection.Drop(context.Background())
		client.Disconnect(context.Background())
	}

	return collection, cleanup
}

func TestDocumentRepository_CreateDocument(t *testing.T) {
	collection, cleanup := setupTestDocumentDB()
	defer cleanup()

	repo := NewDocumentRepository(collection)

	doc := &models.Document{
		ID:     primitive.NewObjectID().Hex(),
		Name:   "test.txt",
		MIME:   "text/plain",
		File:   true,
		Public: false,
		Owner:  "testuser",
		Grant:  []string{"user1", "user2"},
	}

	err := repo.CreateDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
}

func TestDocumentRepository_GetDocumentByID(t *testing.T) {
	collection, cleanup := setupTestDocumentDB()
	defer cleanup()

	repo := NewDocumentRepository(collection)

	doc := &models.Document{
		ID:     primitive.NewObjectID().Hex(),
		Name:   "test.txt",
		MIME:   "text/plain",
		File:   true,
		Public: false,
		Owner:  "testuser",
		Grant:  []string{"user1"},
	}

	err := repo.CreateDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	retrievedDoc, err := repo.GetDocumentByID(context.Background(), doc.ID)
	if err != nil {
		t.Fatalf("Failed to get document: %v", err)
	}

	if retrievedDoc.Name != doc.Name {
		t.Errorf("Expected name %s, got %s", doc.Name, retrievedDoc.Name)
	}

	_, err = repo.GetDocumentByID(context.Background(), primitive.NewObjectID().Hex())
	if err == nil {
		t.Error("Expected error for non-existing document")
	}
}

func TestDocumentRepository_GetDocuments(t *testing.T) {
	collection, cleanup := setupTestDocumentDB()
	defer cleanup()

	repo := NewDocumentRepository(collection)

	docs := []*models.Document{
		{
			ID:     primitive.NewObjectID().Hex(),
			Name:   "doc1.txt",
			MIME:   "text/plain",
			File:   true,
			Public: false,
			Owner:  "user1",
			Grant:  []string{},
		},
		{
			ID:     primitive.NewObjectID().Hex(),
			Name:   "doc2.txt",
			MIME:   "text/plain",
			File:   true,
			Public: true,
			Owner:  "user2",
			Grant:  []string{},
		},
	}

	for _, doc := range docs {
		err := repo.CreateDocument(context.Background(), doc)
		if err != nil {
			t.Fatalf("Failed to create document: %v", err)
		}
	}

	params := models.QueryParams{
		Token: "test_token",
		Limit: 10,
	}

	documents, err := repo.GetDocuments(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to get documents: %v", err)
	}

	if len(documents) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(documents))
	}

	params.Login = "user1"
	documents, err = repo.GetDocuments(context.Background(), params)
	if err != nil {
		t.Fatalf("Failed to get documents: %v", err)
	}

	if len(documents) != 1 {
		t.Errorf("Expected 1 document for user1, got %d", len(documents))
	}
}

func TestDocumentRepository_CheckDocumentAccess(t *testing.T) {
	collection, cleanup := setupTestDocumentDB()
	defer cleanup()

	repo := NewDocumentRepository(collection)

	doc := &models.Document{
		ID:     primitive.NewObjectID().Hex(),
		Name:   "test.txt",
		MIME:   "text/plain",
		File:   true,
		Public: false,
		Owner:  "owner",
		Grant:  []string{"granted_user"},
	}

	err := repo.CreateDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	hasAccess, err := repo.CheckDocumentAccess(context.Background(), doc.ID, "owner")
	if err != nil {
		t.Fatalf("Failed to check access: %v", err)
	}
	if !hasAccess {
		t.Error("Owner should have access")
	}

	hasAccess, err = repo.CheckDocumentAccess(context.Background(), doc.ID, "granted_user")
	if err != nil {
		t.Fatalf("Failed to check access: %v", err)
	}
	if !hasAccess {
		t.Error("Granted user should have access")
	}

	hasAccess, err = repo.CheckDocumentAccess(context.Background(), doc.ID, "denied_user")
	if err != nil {
		t.Fatalf("Failed to check access: %v", err)
	}
	if hasAccess {
		t.Error("Denied user should not have access")
	}
}
