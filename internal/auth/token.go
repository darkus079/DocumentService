package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"documentservice/internal/models"
	"documentservice/internal/validation"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type TokenManager struct {
	usersCollection *mongo.Collection
	adminToken      string
}

func NewTokenManager(usersCollection *mongo.Collection, adminToken string) *TokenManager {
	return &TokenManager{
		usersCollection: usersCollection,
		adminToken:      adminToken,
	}
}

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

	opts := options.Replace().SetUpsert(true)
	filter := bson.M{"_id": login}
	_, err := tm.usersCollection.ReplaceOne(ctx, filter, user, opts)
	if err != nil {
		return "", err
	}

	return token, nil
}

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

func (tm *TokenManager) RegisterUser(adminToken, login, password string) (*models.RegisterResponse, error) {
	if adminToken != tm.adminToken {
		return nil, fmt.Errorf("invalid admin token")
	}

	if err := validation.ValidateLogin(login); err != nil {
		return nil, err
	}

	if err := validation.ValidatePassword(password); err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	user := models.User{
		Login:        login,
		Token:        token,
		PasswordHash: string(passwordHash),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Replace().SetUpsert(true)
	filter := bson.M{"_id": login}
	_, err = tm.usersCollection.ReplaceOne(ctx, filter, user, opts)
	if err != nil {
		return nil, err
	}

	response := &models.RegisterResponse{}
	response.Response.Login = login
	return response, nil
}

func (tm *TokenManager) AuthenticateUser(login, password string) (*models.AuthResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	filter := bson.M{"_id": login}
	err := tm.usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	response := &models.AuthResponse{}
	response.Response.Token = user.Token
	return response, nil
}

func (tm *TokenManager) InvalidateToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"token": token}
	update := bson.M{"$unset": bson.M{"token": ""}}
	_, err := tm.usersCollection.UpdateOne(ctx, filter, update)
	return err
}
