// Package app TUI 应用主入口
// 遵循 Bubble Tea 架构：所有状态变更必须通过 Update 循环
package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StreamingEvent 表示流式输出中的一个事件
type StreamingEvent struct {
	Timestamp time.Time         `json:"timestamp"`
	EventType string            `json:"event_type"` // "stream_msg", "tool_start", "tool_result", "render", "error"
	Data      map[string]any    `json:"data"`
}

// StreamingTracker 流式输出追踪器
// 用于调试流式输出字符丢失问题
type StreamingTracker struct {
	file     *os.File
	encoder  *json.Encoder
	mu       sync.Mutex
	enabled  bool
	filename string
}

// NewStreamingTracker 创建新的流式输出追踪器
// 如果 debug 为 false，返回一个禁用的 tracker
func NewStreamingTracker(debug bool) (*StreamingTracker, error) {
	if !debug {
		return &StreamingTracker{enabled: false}, nil
	}

	// 创建追踪文件目录
	tmpDir := os.TempDir()
	filename := filepath.Join(tmpDir, fmt.Sprintf("qodercli-streaming-trace-%d.json", time.Now().Unix()))

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

// LogStreamMsg 记录流式消息事件
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
			"content":                  content,
			"content_bytes":            len(content),
			"content_runes":            len([]rune(content)),
			"persisted_length":         persistedLen,
			"new_content":              newContent,
			"new_content_bytes":        len(newContent),
			"new_content_runes":        len([]rune(newContent)),
			"streaming_text_length":    streamingTextLen,
		},
	}
	t.encoder.Encode(event)
}

// LogToolStart 记录工具调用开始事件
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
			"content":                  content,
			"content_bytes":            len(content),
			"content_runes":            len([]rune(content)),
			"persisted_len_before":     persistedLenBefore,
			"persisted_len_after":      persistedLenAfter,
		},
	}
	t.encoder.Encode(event)
}

// LogRender 记录 Markdown 渲染事件
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
			"input":           input,
			"input_bytes":     len(input),
			"input_runes":     len([]rune(input)),
			"output":          output,
			"output_bytes":    len(output),
			"output_runes":    len([]rune(output)),
			"renderer_style":  rendererStyle,
		},
	}
	t.encoder.Encode(event)
}

// LogError 记录错误事件
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

// Close 关闭追踪器并返回文件路径
func (t *StreamingTracker) Close() (string, error) {
	if !t.enabled {
		return "", nil
	}

	filename := t.filename
	err := t.file.Close()
	return filename, err
}

// GetFilename 返回追踪文件路径
func (t *StreamingTracker) GetFilename() string {
	return t.filename
}
