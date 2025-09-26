package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"documentservice/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TokenManager handles token operations
type TokenManager struct {
	usersCollection *mongo.Collection
}

// NewTokenManager creates a new token manager
func NewTokenManager(usersCollection *mongo.Collection) *TokenManager {
	return &TokenManager{
		usersCollection: usersCollection,
	}
}

// GenerateToken generates a new token for a user
func (tm *TokenManager) GenerateToken(login string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	user := models.User{
		Login: login,
		Token: token,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Upsert user with new token
	opts := options.Replace().SetUpsert(true)
	filter := bson.M{"_id": login}
	_, err := tm.usersCollection.ReplaceOne(ctx, filter, user, opts)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateToken validates a token and returns the user login
func (tm *TokenManager) ValidateToken(token string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	filter := bson.M{"token": token}
	err := tm.usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("invalid token")
		}
		return "", err
	}

	return user.Login, nil
}

// GetUserByToken gets user information by token
func (tm *TokenManager) GetUserByToken(token string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	filter := bson.M{"token": token}
	err := tm.usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return &user, nil
}
