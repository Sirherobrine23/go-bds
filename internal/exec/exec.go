package exec

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"sirherobrine23.org/Minecraft-Server/go-bds/internal/ioextends"
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

type ServerOptions struct {
	Cwd         string            `json:"cwd"`       // Folder to run server
	Arguments   []string          `json:"arguments"` // Server command and arguments
	Environment map[string]string `json:"env"`       // Process env
}

type Server struct {
	ProcessState *os.ProcessState // Process state
	Process      *os.Process      // Process
	Stdin        io.WriteCloser   // Write to stdin stream
	Stdlog       *ioextends.Piped // Log stdout and stderr
}

func (opts *ServerOptions) Run() (Server, error) {
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
	piped := ioextends.ReadPipe()
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

func Run(opts ServerOptions) (Server, error) {
	return opts.Run()
}

// Write to stdin
func (w *Server) Writer(p []byte) (n int, err error) {
	return w.Stdin.Write(p)
}

func QuickRun(cmd string, arguments ...string) (string, error) {
	ex := exec.Command(cmd, arguments...)
	run, err := ex.Output()
	if err != nil {
		return "", err
	}
	return string(run), nil
}
