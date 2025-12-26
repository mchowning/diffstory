package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/review"
	"github.com/mchowning/diffguide/internal/storage"
)

type Server struct {
	reviewService *review.Service
	server        *http.Server
	listener      net.Listener
	verbose       bool
}

func New(store *storage.Store, port string, verbose bool) (*Server, error) {
	mux := http.NewServeMux()

	s := &Server{
		reviewService: review.NewService(store),
		verbose:       verbose,
		server: &http.Server{
			Addr:              "127.0.0.1:" + port,
			Handler:           mux,
			ReadHeaderTimeout: 2 * time.Second,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      5 * time.Second,
			IdleTimeout:       30 * time.Second,
		},
	}

	mux.HandleFunc("/review", s.handleReview)

	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind: %w", err)
	}
	s.listener = ln

	return s, nil
}

func (s *Server) Port() string {
	addr := s.listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("%d", addr.Port)
}

func (s *Server) Run() error {
	if s.verbose {
		log.Printf("Server listening on http://127.0.0.1:%s", s.Port())
	}
	return s.server.Serve(s.listener)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// maxRequestBody limits request body size to 10MB to prevent DoS
const maxRequestBody = 10 * 1024 * 1024

func (s *Server) handleReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	defer r.Body.Close()

	var reviewData model.Review
	if err := json.NewDecoder(r.Body).Decode(&reviewData); err != nil {
		// Check for oversized body using proper type assertion
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "Request body too large (max 10MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err := s.reviewService.Submit(r.Context(), reviewData)
	if err != nil {
		if errors.Is(err, review.ErrMissingWorkingDirectory) {
			http.Error(w, "Missing workingDirectory field", http.StatusBadRequest)
			return
		}
		if errors.Is(err, review.ErrInvalidWorkingDirectory) {
			http.Error(w, "Invalid workingDirectory: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Storage errors
		http.Error(w, "Failed to store review: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if s.verbose {
		log.Printf("Stored review for %s: %s (%d sections)",
			reviewData.WorkingDirectory, reviewData.Title, reviewData.SectionCount())
	}

	w.WriteHeader(http.StatusOK)
}
