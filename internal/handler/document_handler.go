package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"documentservice/internal/auth"
	"documentservice/internal/models"
	"documentservice/internal/service"
)

type DocumentHandler struct {
	docService service.DocumentServiceInterface
	tokenMgr   *auth.TokenManager
}

func NewDocumentHandler(docService service.DocumentServiceInterface, tokenMgr *auth.TokenManager) *DocumentHandler {
	return &DocumentHandler{
		docService: docService,
		tokenMgr:   tokenMgr,
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

func (dh *DocumentHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dh.writeErrorResponse(w, 400, "Invalid request format")
		return
	}

	response, err := dh.tokenMgr.RegisterUser(req.Token, req.Login, req.Pswd)
	if err != nil {
		dh.writeErrorResponse(w, 400, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}

func (dh *DocumentHandler) Auth(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dh.writeErrorResponse(w, 400, "Invalid request format")
		return
	}

	response, err := dh.tokenMgr.AuthenticateUser(req.Login, req.Pswd)
	if err != nil {
		dh.writeErrorResponse(w, 401, "Invalid credentials")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}

func (dh *DocumentHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Path[len("/api/auth/"):]
	if token == "" {
		dh.writeErrorResponse(w, 400, "Token is required")
		return
	}

	err := dh.tokenMgr.InvalidateToken(token)
	if err != nil {
		dh.writeErrorResponse(w, 500, "Internal server error")
		return
	}

	response := map[string]interface{}{
		"response": map[string]bool{
			token: true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}

func (dh *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		dh.writeErrorResponse(w, 400, "Failed to parse multipart form")
		return
	}

	metaStr := r.FormValue("meta")
	if metaStr == "" {
		dh.writeErrorResponse(w, 400, "Meta parameter is required")
		return
	}

	var meta models.DocumentMeta
	if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
		dh.writeErrorResponse(w, 400, "Invalid meta format")
		return
	}

	var jsonData interface{}
	if jsonStr := r.FormValue("json"); jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
			dh.writeErrorResponse(w, 400, "Invalid json format")
			return
		}
	}

	var fileContent []byte
	if meta.File {
		file, _, err := r.FormFile("file")
		if err != nil {
			dh.writeErrorResponse(w, 400, "File is required for file documents")
			return
		}
		defer file.Close()

		fileContent, err = io.ReadAll(file)
		if err != nil {
			dh.writeErrorResponse(w, 500, "Failed to read file")
			return
		}
	}

	response, err := dh.docService.UploadDocument(r.Context(), meta, jsonData, fileContent)
	if err != nil {
		dh.writeErrorResponse(w, 400, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}

func (dh *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
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

	response, err := dh.docService.DeleteDocument(r.Context(), docID, token)
	if err != nil {
		dh.writeErrorResponse(w, 400, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
}

func (dh *DocumentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := models.NewResponseMessage("Service is healthy")
	dh.writeResponse(w, response)
}
