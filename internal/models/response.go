package models

type APIResponse struct {
	Error    *ErrorResponse `json:"error,omitempty"`
	Response *string        `json:"response,omitempty"`
	Data     interface{}    `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Data: data,
	}
}

func NewErrorResponse(code int, text string) *APIResponse {
	return &APIResponse{
		Error: &ErrorResponse{
			Code: code,
			Text: text,
		},
	}
}

func NewResponseMessage(message string) *APIResponse {
	return &APIResponse{
		Response: &message,
	}
}
