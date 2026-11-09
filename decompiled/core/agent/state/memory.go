package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// SessionMemory Session memory storage - used to store temporary data such as Todo lists
// Persisted to disk at: ~/.qodercli-reverse/sessions/<session-id>.json
type SessionMemory struct {
	mu      sync.RWMutex
	data    map[string]json.RawMessage
	session string // current session name
}

// NewSessionMemory Creates a new session memory
func NewSessionMemory() *SessionMemory {
	return &SessionMemory{
		data: make(map[string]json.RawMessage),
		session: "default",
	}
}

// SessionPath returns the file path for a session.
func SessionPath(session string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	sessionsDir := filepath.Join(home, ".qodercli-reverse", "sessions")
	os.MkdirAll(sessionsDir, 0755)
	return filepath.Join(sessionsDir, regexp.MustCompile(`\.json$`).ReplaceAllString(session, ".json"))
}

// Get Retrieves data.
func (m *SessionMemory) Get(key string) (json.RawMessage, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.data[key]
	return value, ok
}

// Set Stores data and persist to disk.
func (m *SessionMemory) Set(key string, value json.RawMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value

	// Persist to disk
	if err := m.persist(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to persist session: %v\n", err)
	}
}

// Delete Deletes data and persist.
func (m *SessionMemory) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)

	// Persist to disk
	if err := m.persist(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to persist session: %v\n", err)
	}
}

// Clear Clears all data and persist.
func (m *SessionMemory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]json.RawMessage)

	// Persist to disk
	if err := m.persist(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to persist session: %v\n", err)
	}
}

// Keys Returns all keys.
func (m *SessionMemory) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// loadFromDisk loads data from disk into memory.
func (m *SessionMemory) loadFromDisk() error {
	path := SessionPath(m.session)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no saved session yet
		}
		return err
	}

	var dataMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return err
	}

	m.mu.Lock()
	m.data = dataMap
	m.mu.Unlock()

	return nil
}

// persist writes the current memory state to disk.
func (m *SessionMemory) persist() error {
	path := SessionPath(m.session)

	// Convert data to JSON
	data, err := json.Marshal(m.data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ListSessions lists all saved sessions.
func ListSessions() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sessionsDir := filepath.Join(home, ".qodercli-reverse", "sessions")
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, err
	}

	var sessions []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		// Remove .json extension for display
		name := strings.TrimSuffix(entry.Name(), ".json")
		sessions = append(sessions, name)
	}

	return sessions, nil
}

// LoadSession loads a specific session by name.
func (m *SessionMemory) LoadSession(session string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sessionsDir := filepath.Join(home, ".qodercli-reverse", "sessions")
	path := filepath.Join(sessionsDir, strings.TrimSuffix(session, ".json")+".json")

	// Check if session file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("session %q not found", session)
	}

	// Read and validate the session file
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Simple validation: file should contain valid JSON
	if !json.Valid(data) {
		return fmt.Errorf("session %q is not valid JSON", session)
	}

	m.mu.Lock()
	m.session = session
	m.data = make(map[string]json.RawMessage)

	// Try to unmarshal existing data
	if len(data) > 0 {
		if err := json.Unmarshal(data, &m.data); err != nil {
			// If unmarshal fails, start with empty data
			m.data = make(map[string]json.RawMessage)
		}
	}

	m.mu.Unlock()

	return nil
}

// GetSession returns the current session name.
func (m *SessionMemory) GetSession() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.session
}

// SaveSessionAs saves the current session under a new name.
func (m *SessionMemory) SaveSessionAs(newSession string) error {
	// Save current data under the new session name
	path := SessionPath(newSession)

	// Convert data to JSON
	data, err := json.Marshal(m.data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// SessionStats session statistics.
type SessionStats struct {
	MessageCount       int
	ToolCallCount      int
	AssistantReplies   int
	CreatedAt          string
	LastAccessedAt     string
}

// String returns a string representation of the session memory.
func (m *SessionMemory) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := m.Keys()
	return fmt.Sprintf("SessionMemory: %d keys, session: %s", len(keys), m.session)
}