package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Server struct {
	// Server Arguments
	Arguments []string

	// File to save log server
	StdoutFile, StderrFile string

	Cwd string

	// Call function on new line parse to stderr or stdout
	ProcessLine func(line string) error
}

type ServerRun struct {
	Process   *os.Process
	WriteLine func(args ...string) error
	GetStdout func() (io.Reader, error)
	GetStderr func() (io.Reader, error)
}

func ExecCommand(execOptions Server) (ServerRun, error) {
	if len(execOptions.Arguments) == 0 {
		return ServerRun{}, fmt.Errorf("arguments require one command")
	}

	run := exec.Command(execOptions.Arguments[0], execOptions.Arguments[:1]...)
	run.Path = execOptions.Cwd

	stdoutFile, err := os.Create(execOptions.StdoutFile)
	if err != nil {
		return ServerRun{}, err
	}

	stderrFile, err := os.Create(execOptions.StderrFile)
	if err != nil {
		return ServerRun{}, err
	}

	run.Stdout = stdoutFile
	run.Stderr = stderrFile

	err = run.Run()
	if err != nil {
		return ServerRun{}, err
	}

	return ServerRun{
		Process: run.Process,
		WriteLine: func(args ...string) error {
			_, err := run.Stdin.Read([]byte(strings.Join(args, " ")))
			return err
		},
		GetStdout: func() (io.Reader, error) {
			return os.Open(execOptions.StdoutFile)
		},
		GetStderr: func() (io.Reader, error) {
			return os.Open(execOptions.StderrFile)
		},
	}, nil
}
