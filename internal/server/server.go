package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"documentservice/internal/auth"
	"documentservice/internal/config"
	"documentservice/internal/database"
	"documentservice/internal/handler"
	"documentservice/internal/repository"
	"documentservice/internal/service"

	"github.com/gorilla/mux"
)

type Server struct {
	config *config.Config
	server *http.Server
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
	}
}

func (s *Server) Start() error {
	mongoDB, err := database.NewMongoDB(s.config)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer mongoDB.Close()

	usersCollection := mongoDB.Database.Collection("users")
	documentsCollection := mongoDB.Database.Collection("documents")

	tokenMgr := auth.NewTokenManager(usersCollection, s.config.AdminToken)
	docRepo := repository.NewDocumentRepository(documentsCollection)
	docService := service.NewDocumentService(docRepo, tokenMgr, s.config.MaxFileSize)

	docHandler := handler.NewDocumentHandler(docService, tokenMgr)

	router := s.setupRoutes(docHandler)

	s.server = &http.Server{
		Addr:         ":" + s.config.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", s.config.ServerPort)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server exited")
	return nil
}

func (s *Server) setupRoutes(docHandler *handler.DocumentHandler) *mux.Router {
	router := mux.NewRouter()

	api := router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/register", docHandler.Register).Methods("POST")
	api.HandleFunc("/auth", docHandler.Auth).Methods("POST")
	api.HandleFunc("/auth/{token}", docHandler.Logout).Methods("DELETE")
	api.HandleFunc("/docs", docHandler.GetDocuments).Methods("GET", "HEAD")
	api.HandleFunc("/docs", docHandler.UploadDocument).Methods("POST")
	api.HandleFunc("/docs/{id}", docHandler.GetDocument).Methods("GET", "HEAD")
	api.HandleFunc("/docs/{id}", docHandler.DeleteDocument).Methods("DELETE")

	router.HandleFunc("/health", docHandler.HealthCheck).Methods("GET")

	router.Use(s.loggingMiddleware)

	return router
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
