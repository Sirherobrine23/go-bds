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

type ServerRun struct {
	Arguments      []string          // Command and Arguments to run server
	Environment    map[string]string // Process env
	Cwd            string            // Folder to run server
	Stdin      io.WriteCloser    // Stdin to write and communicate to server
	Stdout, Stderr io.ReadCloser     // stderr and stdout log
	process        *os.Process       // Server process
}

func (w *ServerRun) Kill() error {
	return w.process.Kill()
}

func (w *ServerRun) Pid() int {
	return w.process.Pid
}

func (w *ServerRun) Wait() (*os.ProcessState, error) {
	return w.process.Wait()
}

func (Server *ServerRun) Start() error {
	var cmd *exec.Cmd
	if len(Server.Arguments) == 0 {
		return ErrNoCommand
	} else if len(Server.Arguments) == 1 {
		cmd = exec.Command(Server.Arguments[0])
	} else {
		cmd = exec.Command(Server.Arguments[0], Server.Arguments[1:]...)
	}

	// Process cwd
	if len(Server.Cwd) > 0 {
		cmd.Dir = Server.Cwd
	}

	// Copy current envs to process
	for envKey, envValue := range Server.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envKey, envValue))
	}

	// Pipes
	var err error
	if Server.Stdin, err = cmd.StdinPipe(); err != nil {
		return fmt.Errorf("could not get stdin pipe: %v", err)
	} else if Server.Stderr, err = cmd.StderrPipe(); err != nil {
		return fmt.Errorf("could not get stderr pipe: %v", err)
	} else if Server.Stdout, err = cmd.StdoutPipe(); err != nil {
		return fmt.Errorf("could not get stdout pipe: %v", err)
	}

	// Start server
	if err = cmd.Start(); err != nil {
		return err
	}

	// Pipe process to server struct
	Server.process = cmd.Process

	return nil
}
