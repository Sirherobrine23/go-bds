package exec

import (
	"errors"
	"io"
)

var (
	ErrRunning   error = errors.New("process running")         // Process started
	ErrNoRunning error = errors.New("process nothing running") // Process not started or not running
)

type ProcExec struct {
	Arguments   []string          // command and arguments
	Cwd         string            // Workdir path
	Environment map[string]string // Envs to add to process
}

// Universal process struct
type Proc interface {
	Start(options ProcExec) error       // Start process in background
	Kill() error                        // Kill process with SIGKILL
	Wait() error                        // Wait process
	Close() error                       // Send ctrl + c (SIGINT) and wait process end
	Write(p []byte) (int, error)        // Write to stdin
	ExitCode() (int64, error)           // Get process exit
	StdinFork() (io.WriteCloser, error) // Create stdin fork to write
	StdoutFork() (io.ReadCloser, error) // Create stdout fork to read log
	StderrFork() (io.ReadCloser, error) // Create stderr fork to read log
}
