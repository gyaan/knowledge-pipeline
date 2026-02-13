package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gyaan/knowledge-pipeline/internal/rag"
	"github.com/gyaan/knowledge-pipeline/internal/session"
	"github.com/gyaan/knowledge-pipeline/internal/vectordb"
	"github.com/gyaan/knowledge-pipeline/pkg/models"
)

type Handlers struct {
	ragChain       *rag.Chain
	sessionManager *session.Manager
	vectorDB       *vectordb.MemoryVectorDB
}

func NewHandlers(ragChain *rag.Chain, sessionManager *session.Manager, vectorDB *vectordb.MemoryVectorDB) *Handlers {
	return &Handlers{
		ragChain:       ragChain,
		sessionManager: sessionManager,
		vectorDB:       vectorDB,
	}
}

// ChatHandler handles chat requests.
func (h *Handlers) ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	sess := h.sessionManager.GetOrCreate(req.SessionID)

	// Add user message to session
	userMsg := models.ChatMessage{
		Role:      "user",
		Content:   req.Query,
		Timestamp: time.Now(),
	}
	h.sessionManager.AddMessage(sess.ID, userMsg)

	// Get chat history (exclude the message we just added)
	history := sess.Messages[:len(sess.Messages)-1]

	response, err := h.ragChain.Query(r.Context(), req.Query, history)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing query: %v", err), http.StatusInternalServerError)
		return
	}

	// Add assistant message to session
	assistantMsg := models.ChatMessage{
		Role:      "assistant",
		Content:   response.Answer,
		Timestamp: time.Now(),
	}
	h.sessionManager.AddMessage(sess.ID, assistantMsg)

	response.SessionID = sess.ID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthHandler returns health status.
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "healthy",
		"active_sessions": h.sessionManager.Count(),
		"documents":       h.vectorDB.Count(),
	})
}
