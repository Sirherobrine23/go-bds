package mclog

import (
	"errors"
	"fmt"
	"time"

	"sirherobrine23.com.br/go-bds/go-bds/logs"
)

var (
	MclogsApi  string = "https://api.mclo.gs"
	MclogsBase string = "https://mclo.gs"

	ErrNoExists error = errors.New("log no exists")      // id request not exists
	ErrNoId     error = errors.New("require mclo.gs id") // Require uploaded log to view
)

const (
	DefaultMaxLength int64 = 10485760
	DefaultMaxLines  int64 = 25000

	LogUnknown LogLevel = "unknown"
	LogInfo    LogLevel = "information"
	LogProblem LogLevel = "problems"
	LogWarn    LogLevel = "warning"
)

type Limits struct {
	StorageTime time.Duration `json:"storageTime"` // The duration in seconds that a log is stored for after the last view.
	MaxLength   int64         `json:"maxLength"`   // Maximum file length in bytes. Logs over this limit will be truncated to this length.
	MaxLines    int64         `json:"maxLines"`    // Maximum number of lines. Additional lines will be removed.
}

// Stands log levels
type LogLevel string

type MclogResponseStatus struct {
	Success      bool   `json:"success"`         // Request return is processed request
	ErrorMessage string `json:"error,omitempty"` // Real error question in bad request's
	Id           string `json:"id,omitempty"`    // If post log file return id if processed
}

type EntryLine struct {
	Numbers int    `json:"number"`
	Content string `json:"content"`
	Label   string `json:"label,omitempty"`
}

type AnalysisEntry struct {
	Level     int         `json:"level"`
	Prefix    string      `json:"prefix"`
	EntryTime time.Time   `json:"time"`
	Lines     []EntryLine `json:"lines"`
}

type InsightsAnalysis struct {
	Label    string        `json:"label"`
	Value    string        `json:"value"`
	Message  string        `json:"message"`
	Counter  int           `json:"counter"`
	External any           `json:"external,omitempty"`
	Entry    AnalysisEntry `json:"level"`
}

type Insights struct {
	ID       string                           `json:"id"`
	Version  string                           `json:"version"`
	Title    string                           `json:"title"`
	Type     string                           `json:"type,omitempty"`
	Name     string                           `json:"name,omitempty"`
	Analysis map[LogLevel][]*InsightsAnalysis `json:"analysis"`
}

func ConvertLogs(id string, log logs.Log) Insights {
	var insight Insights
	insight.Type = "server"
	insight.ID = id

	serverInfo := log.Server()
	insight.Version = serverInfo.Version
	insight.Name = serverInfo.Platform
	insight.Title = fmt.Sprintf("%s - %s", serverInfo.Platform, serverInfo.Version)

	insight.Analysis = map[LogLevel][]*InsightsAnalysis{}
	insight.Analysis[LogInfo] = append(insight.Analysis[LogInfo], &InsightsAnalysis{
		Label: "Started time",
		Value: serverInfo.Started.Format(time.RFC3339),
	})

	for _, port := range serverInfo.Ports {
		insight.Analysis[LogInfo] = append(insight.Analysis[LogInfo], &InsightsAnalysis{
			Label:   "Port listen",
			Message: port.From,
			Value:   port.AddrPort.String(),
		})
	}

	for _, err := range log.Errors() {
		insight.Analysis[LogProblem] = append(insight.Analysis[LogProblem], &InsightsAnalysis{
			Value: err.Error(),
		})
	}
	for _, err := range log.Warnings() {
		insight.Analysis[LogWarn] = append(insight.Analysis[LogWarn], &InsightsAnalysis{
			Value: err.Error(),
		})
	}

	return insight
}
