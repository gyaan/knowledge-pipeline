package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/gyaan/knowledge-pipeline/pkg/models"
)

type Manager struct {
	sessions map[string]*models.Session
	mu       sync.RWMutex
	ttl      time.Duration
	done     chan struct{}
}

func NewManager(ttlHours int) *Manager {
	m := &Manager{
		sessions: make(map[string]*models.Session),
		ttl:      time.Duration(ttlHours) * time.Hour,
		done:     make(chan struct{}),
	}
	go m.cleanup()
	return m
}

// Stop stops the background cleanup goroutine.
func (m *Manager) Stop() {
	close(m.done)
}

// GetOrCreate returns the session for the given ID, creating one if the ID is empty or unknown.
func (m *Manager) GetOrCreate(sessionID string) *models.Session {
	if sessionID != "" {
		m.mu.RLock()
		sess, ok := m.sessions[sessionID]
		m.mu.RUnlock()
		if ok {
			return sess
		}
	}

	if sessionID == "" {
		sessionID = generateSessionID()
	}
	return m.Create(sessionID)
}

// Create creates a new session with the given ID.
func (m *Manager) Create(sessionID string) *models.Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &models.Session{
		ID:        sessionID,
		Messages:  make([]models.ChatMessage, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.sessions[sessionID] = session
	return session
}

// Get retrieves a session by ID.
func (m *Manager) Get(sessionID string) (*models.Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessions[sessionID]
	return session, exists
}

// AddMessage adds a message to a session.
func (m *Manager) AddMessage(sessionID string, message models.ChatMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		session.Messages = append(session.Messages, message)
		session.UpdatedAt = time.Now()
	}
}

// Delete deletes a session.
func (m *Manager) Delete(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// Count returns the number of active sessions.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

func (m *Manager) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for id, session := range m.sessions {
				if now.Sub(session.UpdatedAt) > m.ttl {
					delete(m.sessions, id)
				}
			}
			m.mu.Unlock()
		}
	}
}

func generateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID
		return "sess_" + hex.EncodeToString([]byte(time.Now().String()))
	}
	return "sess_" + hex.EncodeToString(b)
}
