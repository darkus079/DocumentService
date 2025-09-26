package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"documentservice/internal/models"
	"documentservice/internal/service"
)

type DocumentHandler struct {
	docService service.DocumentServiceInterface
}

func NewDocumentHandler(docService service.DocumentServiceInterface) *DocumentHandler {
	return &DocumentHandler{
		docService: docService,
	}
}

func (dh *DocumentHandler) GetDocuments(w http.ResponseWriter, r *http.Request) {
	params := models.QueryParams{
		Token: r.URL.Query().Get("token"),
		Login: r.URL.Query().Get("login"),
		Key:   r.URL.Query().Get("key"),
		Value: r.URL.Query().Get("value"),
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if params.Token == "" {
		dh.writeErrorResponse(w, 400, "Token is required")
		return
	}

	response, err := dh.docService.GetDocuments(r.Context(), params)
	if err != nil {
		dh.writeErrorResponse(w, 500, "Internal server error")
		return
	}

	if response.Error != nil {
		dh.writeErrorResponse(w, response.Error.Code, response.Error.Text)
		return
	}

	dh.writeResponse(w, response)
}

func (dh *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	docID := r.URL.Path[len("/api/docs/"):]
	if docID == "" {
		dh.writeErrorResponse(w, 400, "Document ID is required")
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		dh.writeErrorResponse(w, 400, "Token is required")
		return
	}

	response, httpResp, err := dh.docService.GetDocument(r.Context(), docID, token)
	if err != nil {
		dh.writeErrorResponse(w, 500, "Internal server error")
		return
	}

	if response.Error != nil {
		dh.writeErrorResponse(w, response.Error.Code, response.Error.Text)
		return
	}

	if httpResp != nil {
		for key, values := range httpResp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(httpResp.StatusCode)

		if response.Data != nil {
			if doc, ok := response.Data.(*models.Document); ok && len(doc.Content) > 0 {
				w.Write(doc.Content)
			}
		}
		return
	}

	dh.writeResponse(w, response)
}

func (dh *DocumentHandler) writeResponse(w http.ResponseWriter, response *models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (dh *DocumentHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := models.NewErrorResponse(statusCode, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

func (dh *DocumentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := models.NewResponseMessage("Service is healthy")
	dh.writeResponse(w, response)
}
