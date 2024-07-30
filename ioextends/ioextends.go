package ioextends

import (
	"io"
	"os"
)

type Piped struct {
	closed bool
	writer []io.WriteCloser
}

// create new ReadCloser
func (w *Piped) NewPipe() (io.ReadCloser, error) {
	r, write, err := os.Pipe()
	if err == nil {
		w.writer = append(w.writer, write)
	}
	return r, err
}

// Add existing io.WriteCloser to list
func (w *Piped) AddWriter(wr io.WriteCloser) error {
	if w.closed {
		return io.EOF
	}
	w.writer = append(w.writer, wr)
	return nil
}

// Close all writes, if io.WriteCloser.Close() returns an error, the loop will be stopped and this error will be returned, if no error occurs, return an io.EOF
func (w *Piped) Close() error {
	w.closed = true
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

// Write bytes to pipeds streams
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
	return &Piped{false, []io.WriteCloser{}}
}
