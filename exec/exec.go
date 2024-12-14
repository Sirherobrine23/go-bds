package exec

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrRunning   error = errors.New("process running")         // Process started
	ErrNoRunning error = errors.New("process nothing running") // Process not started or not running
)

type Env map[string]string

func (env Env) ToSlice() []string {
	data := []string{}
	for key, value := range env {
		data = append(data, fmt.Sprintf("%s=%s", key, value))
	}
	return data
}

// Generic struct to start Process
type ProcExec struct {
	Arguments   []string          // command and arguments
	Cwd         string            // Workdir path
	Environment Env // Envs to add to process
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

// Write to many streamings and if closed remove from list
type MultiWrite struct {
	Std    []io.Writer
	Closed bool
}

func (p *MultiWrite) AddNewWriter(w io.Writer) {
	if !p.Closed {
		p.Std = append(p.Std, w)
	}
}

func (p *MultiWrite) Close() error {
	p.Closed = true
	return nil
}

func (p *MultiWrite) Write(w []byte) (int, error) {
	if p.Closed {
		return 0, io.EOF
	}

	for indexWriter := range p.Std {
		if p.Std[indexWriter] == nil {
			continue
		}
		switch _, err := p.Std[indexWriter].Write(w); err {
		case nil:
			continue
		case io.EOF, io.ErrUnexpectedEOF:
			p.Std[indexWriter] = nil
		default:
			return 0, err
		}
	}

	for writeIndex := 0; writeIndex < len(p.Std); writeIndex++ {
		if p.Std[writeIndex] != nil {
			continue
		} else if writeIndex == len(p.Std) {
			p.Std = p.Std[writeIndex-1:]
		} else if writeIndex == 0 {
			p.Std = p.Std[1:]
		} else {
			p.Std = append(p.Std[writeIndex:], p.Std[writeIndex+1:]...)
		}
		writeIndex -= 2
	}

	return len(w), nil
}
