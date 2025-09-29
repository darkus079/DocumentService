package validation

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "Password123!",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Pass1!",
			wantErr:  true,
		},
		{
			name:     "no uppercase",
			password: "password123!",
			wantErr:  true,
		},
		{
			name:     "no lowercase",
			password: "PASSWORD123!",
			wantErr:  true,
		},
		{
			name:     "no digit",
			password: "Password!",
			wantErr:  true,
		},
		{
			name:     "no special char",
			password: "Password123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		wantErr bool
	}{
		{
			name:    "valid login",
			login:   "testuser123",
			wantErr: false,
		},
		{
			name:    "too short",
			login:   "test",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			login:   "test-user!",
			wantErr: true,
		},
		{
			name:    "empty login",
			login:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogin(tt.login)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLogin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
