package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type dynamicWrite struct {
	p []io.Writer
}

func (p *dynamicWrite) Append(w io.Writer) {
	p.p = append(p.p, w)
}

func (p *dynamicWrite) Write(w []byte) (int, error) {
	for indexWriter, wri := range p.p {
		_, err := wri.Write(w[:])
		if err == nil {
			continue
		} else if err == io.EOF {
			if (indexWriter + 1) == len(p.p) {
				p.p = p.p[:indexWriter]
			} else {
				p.p = append(p.p[:indexWriter], p.p[indexWriter+1:]...)
			}
		} else {
			return 0, err
		}
	}
	return len(w), nil
}

// Check if binary exists
func LocalBinExist(name string) bool {
	binpath, err := exec.LookPath(name)
	return err == nil && binpath != ""
}

type Os struct {
	osProc *exec.Cmd

	stdin              io.WriteCloser
	stdout, stderr     io.ReadCloser
	stderrIO, stdoutIO *dynamicWrite
}

func (os *Os) Write(w []byte) (int, error) {
	if os.stdin == nil {
		return 0, ErrNoRunning
	}
	return os.stdin.Write(w)
}

func (os *Os) Wait() error {
	if os.stdin == nil {
		return ErrNoRunning
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
	if w.stdout != nil {
		w.stdout.Close()
	}
	if w.stderr != nil {
		w.stderr.Close()
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
	go io.Copy(cli.stdin, r) // Write to stdin
	return w, nil
}

func (cli *Os) StdoutFork() (io.ReadCloser, error) {
	r, w := io.Pipe()
	cli.stdoutIO.Append(w)
	return r, nil
}

func (cli *Os) StderrFork() (io.ReadCloser, error) {
	r, w := io.Pipe()
	cli.stderrIO.Append(w)
	return r, nil
}

func (w *Os) Start(options ProcExec) error {
	w.osProc = exec.Command(options.Arguments[0], options.Arguments[1:]...)
	w.osProc.Dir = options.Cwd
	for key, value := range options.Environment {
		w.osProc.Env = append(w.osProc.Env, fmt.Sprintf("%s=%s", key, value))
	}

	w.stderrIO = &dynamicWrite{}
	w.stdoutIO = &dynamicWrite{}
	var err error
	if w.stdout, err = w.osProc.StdoutPipe(); err != nil {
		return err
	} else if w.stderr, err = w.osProc.StderrPipe(); err != nil {
		w.stdout.Close()
		return err
	} else if w.stdin, err = w.osProc.StdinPipe(); err != nil {
		w.stdout.Close()
		w.stderr.Close()
		return err
	}
	go io.Copy(w.stdoutIO, w.stdout)
	go io.Copy(w.stderrIO, w.stderr)

	if err := w.osProc.Start(); err != nil {
		return err
	}
	return nil
}
