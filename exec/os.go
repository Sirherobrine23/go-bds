package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

var _ Proc = &Os{}

// Check if binary exists
func LocalBinExist(processConfig ProcExec) bool {
	binpath, err := exec.LookPath(processConfig.Arguments[0])
	return err == nil && binpath != ""
}

type Os struct {
	osProc *exec.Cmd

	stdin          io.WriteCloser
	stdout, stderr *MultiWrite
}

func (os *Os) Write(w []byte) (int, error) {
	if os.stdin == nil {
		return 0, ErrNoRunning
	}
	return os.stdin.Write(w)
}

func (os *Os) Wait() error {
	if os.osProc == nil {
		return ErrNoRunning
	} else if os.osProc.Process != nil {
		state, err := os.osProc.Process.Wait()
		if err == nil && !state.Success() {
			err = &exec.ExitError{ProcessState: state}
		}
		return err
	}
	return os.osProc.Wait()
}

func (w *Os) Kill() error {
	if w.osProc == nil {
		return ErrNoRunning
	}
	return w.osProc.Process.Kill()
}

func (w *Os) Close() error {
	if w.osProc == nil {
		return ErrNoRunning
	}
	if w.stdin != nil {
		w.stdin.Close()
	}
	return w.osProc.Process.Signal(os.Interrupt)
}

func (w *Os) ExitCode() (int64, error) {
	if !w.osProc.ProcessState.Exited() {
		return 0, ErrRunning
	}
	return int64(w.osProc.ProcessState.ExitCode()), nil
}

func (cli *Os) StdinFork() (io.WriteCloser, error) {
	if cli.stdin == nil {
		return nil, ErrNoRunning
	}
	r, w := io.Pipe()
	// Write to stdin
	//nolint:errcheck
	go io.Copy(cli.stdin, r)
	return w, nil
}

func (cli *Os) StdoutFork() (io.ReadCloser, error) {
	if cli.stdout == nil {
		return nil, ErrNoRunning
	}
	r, w := io.Pipe()
	cli.stdout.AddNewWriter(w)
	return r, nil
}

func (cli *Os) StderrFork() (io.ReadCloser, error) {
	if cli.stderr == nil {
		return nil, ErrNoRunning
	}
	r, w := io.Pipe()
	cli.stderr.AddNewWriter(w)
	return r, nil
}

func (w *Os) Start(options ProcExec) error {
	w.osProc = exec.Command(options.Arguments[0], options.Arguments[1:]...)
	w.osProc.Dir = options.Cwd
	for key, value := range options.Environment {
		w.osProc.Env = append(w.osProc.Env, fmt.Sprintf("%s=%s", key, value))
	}

	w.stderr = &MultiWrite{}
	w.osProc.Stderr = w.stderr

	w.stdout = &MultiWrite{}
	w.osProc.Stdout = w.stdout

	var err error
	if w.stdin, err = w.osProc.StdinPipe(); err != nil {
		return err
	}

	if err := w.osProc.Start(); err != nil {
		return err
	}
	return nil
}
