package exec

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var (
	ErrNoCommand = errors.New("no command, require command")
)

func ProgrammExist(Programm string) bool {
	path, err := exec.LookPath(Programm)
	if err != nil {
		return false
	} else if len(path) == 0 {
		return false
	}
	return true
}

type Piped struct {
	writer []io.WriteCloser
}

func (w *Piped) NewPipe() (io.ReadCloser, error) {
	r, write, err := os.Pipe()
	if err == nil {
		w.writer = append(w.writer, write)
	}
	return r, err
}

func (w *Piped) Close() error {
	if len(w.writer) > 0 {
		for _, w := range w.writer {
			err := w.Close()
			if err == nil || err == io.EOF {
				continue
			}
			return err
		}
		w.writer = []io.WriteCloser{}
		return io.EOF
	}

	return io.EOF
}

func (w *Piped) Write(p []byte) (int, error) {
	for indexWriter, wri := range w.writer {
		_, err := wri.Write(p[:])
		if err == nil {
			continue
		} else if err == io.EOF {
			if (indexWriter + 1) == len(w.writer) {
				w.writer = w.writer[:indexWriter]
			} else {
				w.writer = append(w.writer[:indexWriter], w.writer[indexWriter+1:]...)
			}
		} else {
			return 0, err
		}
	}
	return len(p), nil
}

func ReadPipe() *Piped {
	return &Piped{[]io.WriteCloser{}}
}

type ServerOptions struct {
	Cwd         string            `json:"cwd"`       // Folder to run server
	Arguments   []string          `json:"arguments"` // Server command and arguments
	Environment map[string]string `json:"env"`       // Process env
}

type Server struct {
	ProcessState *os.ProcessState // Process state
	Process      *os.Process      // Process
	Stdin        io.WriteCloser   // Write to stdin stream
	Stdlog       *Piped            // Log stdout and stderr
}

func Run(opts ServerOptions) (Server, error) {
	var cmd *exec.Cmd
	var main Server
	if len(opts.Arguments) == 0 {
		return main, ErrNoCommand
	} else if len(opts.Arguments) == 1 {
		cmd = exec.Command(opts.Arguments[0])
	} else {
		cmd = exec.Command(opts.Arguments[0], opts.Arguments[1:]...)
	}

	// make Server struct
	main = Server{}

	// Process cwd
	if len(opts.Cwd) > 0 {
		cmd.Dir = opts.Cwd
	}

	// Copy current envs to process
	for envKey, envValue := range opts.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envKey, envValue))
	}

	var err error
	if main.Stdin, err = cmd.StdinPipe(); err != nil {
		return main, fmt.Errorf("could get pipe to stdin: %v", err)
	}

	// Pipe stderr and stdout
	piped := ReadPipe()
	cmd.Stderr = piped
	cmd.Stdout = piped
	main.Stdlog = piped

	// Start server
	if err = cmd.Start(); err != nil {
		main.Stdin.Close()
		main.Stdlog.Close()
		return main, err
	}

	main.ProcessState = cmd.ProcessState
	main.Process = cmd.Process

	return main, nil
}

// Write to stdin
func (w *Server) Writer(p []byte) (n int, err error) {
	return w.Stdin.Write(p)
}
