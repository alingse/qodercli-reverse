// Package app TUI application entry point
// Follows Bubble Tea architecture: all state changes must go through the Update loop
package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StreamingEvent represents an event in streaming output
type StreamingEvent struct {
	Timestamp time.Time      `json:"timestamp"`
	EventType string         `json:"event_type"` // "stream_msg", "tool_start", "tool_result", "render", "error"
	Data      map[string]any `json:"data"`
}

// StreamingTracker streaming output tracker
// used for debugging streaming output character loss issues
type StreamingTracker struct {
	file     *os.File
	encoder  *json.Encoder
	mu       sync.Mutex
	enabled  bool
	filename string
}

// NewStreamingTracker creates a new streaming output tracker
// if debug is false, returns a disabled tracker
func NewStreamingTracker(debug bool) (*StreamingTracker, error) {
	if !debug {
		return &StreamingTracker{enabled: false}, nil
	}

	// create trace file directory
	tmpDir := os.TempDir()
	filename := filepath.Join(tmpDir, fmt.Sprintf("qodercli-reverse-streaming-trace-%d.json", time.Now().Unix()))

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming tracker file: %w", err)
	}

	return &StreamingTracker{
		file:     file,
		encoder:  json.NewEncoder(file),
		enabled:  true,
		filename: filename,
	}, nil
}

// LogStreamMsg logs a streaming message event
func (t *StreamingTracker) LogStreamMsg(content string, persistedLen int, newContent string, streamingTextLen int) {
	if !t.enabled {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	event := StreamingEvent{
		Timestamp: time.Now(),
		EventType: "stream_msg",
		Data: map[string]interface{}{
			"content":               content,
			"content_bytes":         len(content),
			"content_runes":         len([]rune(content)),
			"persisted_length":      persistedLen,
			"new_content":           newContent,
			"new_content_bytes":     len(newContent),
			"new_content_runes":     len([]rune(newContent)),
			"streaming_text_length": streamingTextLen,
		},
	}
	t.encoder.Encode(event)
}

// LogToolStart logs a tool call start event
func (t *StreamingTracker) LogToolStart(content string, persistedLenBefore, persistedLenAfter int) {
	if !t.enabled {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	event := StreamingEvent{
		Timestamp: time.Now(),
		EventType: "tool_start",
		Data: map[string]interface{}{
			"content":              content,
			"content_bytes":        len(content),
			"content_runes":        len([]rune(content)),
			"persisted_len_before": persistedLenBefore,
			"persisted_len_after":  persistedLenAfter,
		},
	}
	t.encoder.Encode(event)
}

// LogRender logs a Markdown render event
func (t *StreamingTracker) LogRender(input string, output string, rendererStyle string) {
	if !t.enabled {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	event := StreamingEvent{
		Timestamp: time.Now(),
		EventType: "render",
		Data: map[string]interface{}{
			"input":          input,
			"input_bytes":    len(input),
			"input_runes":    len([]rune(input)),
			"output":         output,
			"output_bytes":   len(output),
			"output_runes":   len([]rune(output)),
			"renderer_style": rendererStyle,
		},
	}
	t.encoder.Encode(event)
}

// LogError logs an error event
func (t *StreamingTracker) LogError(err error, context string) {
	if !t.enabled {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	event := StreamingEvent{
		Timestamp: time.Now(),
		EventType: "error",
		Data: map[string]interface{}{
			"error":   err.Error(),
			"context": context,
		},
	}
	t.encoder.Encode(event)
}

// Close closes the tracker and returns the file path
func (t *StreamingTracker) Close() (string, error) {
	if !t.enabled {
		return "", nil
	}

	filename := t.filename
	err := t.file.Close()
	return filename, err
}

// GetFilename returns the trace file path
func (t *StreamingTracker) GetFilename() string {
	return t.filename
}
