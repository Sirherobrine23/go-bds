package mclog

import (
	"errors"
	"time"
)

var (
	MclogsApi  string = "https://api.mclo.gs"
	MclogsBase string = "https://mclo.gs"

	ErrNoExists        error = errors.New("log no exists")               // id request not exists
	ErrNoId            error = errors.New("require mclo.gs id")          // Require uploaded log to view
)

const (
	DefaultMaxLength int64 = 10485760
	DefaultMaxLines  int64 = 25000

	LogUnknown LogLevel = "unknown"
	LogInfo    LogLevel = "information"
	LogProblem LogLevel = "problems"
	LogWarn    LogLevel = "warning"
)

// Stands log levels
type LogLevel string

type MclogResponseStatus struct {
	Success      bool   `json:"success"`         // Request return is processed request
	ErrorMessage string `json:"error,omitempty"` // Real error question in bad request's
	Id           string `json:"id,omitempty"`    // If post log file return id if processed
}

type Limits struct {
	StorageTime time.Duration `json:"storageTime"` // The duration in seconds that a log is stored for after the last view.
	MaxLength   int64         `json:"maxLength"`   // Maximum file length in bytes. Logs over this limit will be truncated to this length.
	MaxLines    int64         `json:"maxLines"`    // Maximum number of lines. Additional lines will be removed.
}

type EntryLine struct {
	Numbers int64  `json:"number"`
	Content string `json:"content"`
}

type AnalysisEntry struct {
	Level     int64       `json:"level"`
	Prefix    string      `json:"prefix"`
	EntryTime time.Time   `json:"time"`
	Lines     []EntryLine `json:"lines"`
}

type InsightsAnalysis struct {
	Label   string        `json:"label"`
	Value   string        `json:"value"`
	Message string        `json:"message"`
	Counter int64         `json:"counter"`
	Entry   AnalysisEntry `json:"level"`
}

type Insights struct {
	ID       string                          `json:"id"`
	Version  string                          `json:"version"`
	Title    string                          `json:"title"`
	Type     string                          `json:"type,omitempty"`
	Name     string                          `json:"name,omitempty"`
	Analysis map[LogLevel][]InsightsAnalysis `json:"analysis"`
}
