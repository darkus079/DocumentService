package models

import (
	"testing"
)

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"test": "value"}
	response := NewSuccessResponse(data)

	if response.Error != nil {
		t.Error("Success response should not have error")
	}
	if response.Response != nil {
		t.Error("Success response should not have response message")
	}
	if response.Data == nil {
		t.Error("Success response should have data")
	}
}

func TestNewErrorResponse(t *testing.T) {
	code := 400
	text := "Bad request"
	response := NewErrorResponse(code, text)

	if response.Error == nil {
		t.Error("Error response should have error")
	}
	if response.Error.Code != code {
		t.Errorf("Expected error code %d, got %d", code, response.Error.Code)
	}
	if response.Error.Text != text {
		t.Errorf("Expected error text %s, got %s", text, response.Error.Text)
	}
	if response.Response != nil {
		t.Error("Error response should not have response message")
	}
	if response.Data != nil {
		t.Error("Error response should not have data")
	}
}

func TestNewResponseMessage(t *testing.T) {
	message := "Operation successful"
	response := NewResponseMessage(message)

	if response.Error != nil {
		t.Error("Response message should not have error")
	}
	if response.Response == nil {
		t.Error("Response message should have response message")
	}
	if *response.Response != message {
		t.Errorf("Expected message %s, got %s", message, *response.Response)
	}
	if response.Data != nil {
		t.Error("Response message should not have data")
	}
}
