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
	stdout, stderr *Writers
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
	r, w := io.Pipe()
	return w, cli.AppendToStdin(r)
}

func (cli *Os) StdoutFork() (io.ReadCloser, error) {
	r, w := io.Pipe()
	return r, cli.AppendToStdout(w)
}

func (cli *Os) StderrFork() (io.ReadCloser, error) {
	r, w := io.Pipe()
	return r, cli.AppendToStderr(w)
}

func (cli *Os) AppendToStdout(w io.Writer) error {
	if cli.stdout == nil {
		cli.stdout = &Writers{}
	}
	cli.stdout.Std = append(cli.stdout.Std, w)
	return nil
}
func (cli *Os) AppendToStderr(w io.Writer) error {
	if cli.stderr == nil {
		cli.stderr = &Writers{}
	}
	cli.stderr.Std = append(cli.stderr.Std, w)
	return nil
}
func (cli *Os) AppendToStdin(r io.Reader) error {
	if cli.stdin == nil {
		return ErrNoRunning
	}
	go io.Copy(cli.stdin, r)
	return nil
}

func (w *Os) Start(options ProcExec) error {
	w.osProc = exec.Command(options.Arguments[0], options.Arguments[1:]...)
	w.osProc.Dir = options.Cwd
	for key, value := range options.Environment {
		w.osProc.Env = append(w.osProc.Env, fmt.Sprintf("%s=%s", key, value))
	}

	if w.stdout == nil {
		w.stdout = &Writers{}
	}
	w.osProc.Stdout = w.stdout

	if w.stderr == nil {
		w.stderr = &Writers{}
	}
	w.osProc.Stderr = w.stderr

	var err error
	if w.stdin, err = w.osProc.StdinPipe(); err != nil {
		return err
	}

	if err := w.osProc.Start(); err != nil {
		return err
	}
	return nil
}
