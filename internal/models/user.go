package models

type User struct {
	Login        string `json:"login" bson:"_id"`
	Token        string `json:"token" bson:"token"`
	PasswordHash string `json:"-" bson:"password_hash"`
}

type Token struct {
	Value  string `json:"value"`
	UserID string `json:"user_id"`
}

type RegisterRequest struct {
	Token string `json:"token"`
	Login string `json:"login"`
	Pswd  string `json:"pswd"`
}

type AuthRequest struct {
	Login string `json:"login"`
	Pswd  string `json:"pswd"`
}

type RegisterResponse struct {
	Response struct {
		Login string `json:"login"`
	} `json:"response"`
}

type AuthResponse struct {
	Response struct {
		Token string `json:"token"`
	} `json:"response"`
}
