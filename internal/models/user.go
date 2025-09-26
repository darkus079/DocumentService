package models

// User represents a user in the system
type User struct {
	Login string `json:"login" bson:"_id"`
	Token string `json:"token" bson:"token"`
}

// Token represents an authentication token
type Token struct {
	Value  string `json:"value"`
	UserID string `json:"user_id"`
}
