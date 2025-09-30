package repository

import (
	"context"
	"fmt"
	"time"

	"documentservice/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocumentRepository struct {
	collection *mongo.Collection
}

func NewDocumentRepository(collection *mongo.Collection) *DocumentRepository {
	return &DocumentRepository{
		collection: collection,
	}
}

func (dr *DocumentRepository) GetDocuments(ctx context.Context, params models.QueryParams) ([]models.Document, error) {
	filter := bson.M{}

	if params.Login != "" {
		filter["owner"] = params.Login
	}

	if params.Key != "" && params.Value != "" {
		filter[params.Key] = params.Value
	}

	limit := int64(100)
	if params.Limit > 0 {
		limit = int64(params.Limit)
	}

	opts := options.Find().
		SetSort(bson.M{
			"name":    1,
			"created": 1,
		}).
		SetLimit(limit)

	cursor, err := dr.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var documents []models.Document
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}

	return documents, nil
}

func (dr *DocumentRepository) GetDocumentByID(ctx context.Context, id string) (*models.Document, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID format")
	}

	filter := bson.M{"_id": objectID}
	var document models.Document
	err = dr.collection.FindOne(ctx, filter).Decode(&document)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("document not found")
		}
		return nil, err
	}

	return &document, nil
}

func (dr *DocumentRepository) CreateDocument(ctx context.Context, doc *models.Document) error {
	doc.Created = time.Now()
	_, err := dr.collection.InsertOne(ctx, doc)
	return err
}

func (dr *DocumentRepository) UpdateDocument(ctx context.Context, doc *models.Document) error {
	filter := bson.M{"_id": doc.ID}
	update := bson.M{"$set": doc}
	_, err := dr.collection.UpdateOne(ctx, filter, update)
	return err
}

func (dr *DocumentRepository) DeleteDocument(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid document ID format")
	}

	filter := bson.M{"_id": objectID}
	_, err = dr.collection.DeleteOne(ctx, filter)
	return err
}

func (dr *DocumentRepository) CheckDocumentAccess(ctx context.Context, docID, userLogin string) (bool, error) {
	objectID, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		return false, fmt.Errorf("invalid document ID format")
	}

	filter := bson.M{
		"_id": objectID,
		"$or": []bson.M{
			{"owner": userLogin},
			{"public": true},
			{"grant": userLogin},
		},
	}

	count, err := dr.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
