package api

import (
	"context"
	"log"
	"net/http"
	"time"
)

type Server struct {
	handlers *Handlers
	port     string
}

func NewServer(handlers *Handlers, port string) *Server {
	return &Server{
		handlers: handlers,
		port:     port,
	}
}

// Start starts the HTTP server with graceful shutdown support.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/chat", s.corsMiddleware(s.handlers.ChatHandler))
	mux.HandleFunc("/api/health", s.corsMiddleware(s.handlers.HealthHandler))

	server := &http.Server{
		Addr:         ":" + s.port,
		Handler:      s.loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		log.Println("Shutting down server...")
		server.Shutdown(shutdownCtx)
	}()

	log.Printf("Server starting on port %s", s.port)
	return server.ListenAndServe()
}

func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.Method, r.RequestURI, r.RemoteAddr, time.Since(start))
	})
}
