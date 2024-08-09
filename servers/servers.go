// Implement global servers Functions to run Server
package servers

import (
	"io"
	"io/fs"
)

type Server interface {
	Update(Version string) error    // Update/install server
	Start() error                   // Start server in backgroud
	Close() error                   // Clear process and/or stop server
	Stdout() (io.ReadCloser, error) // Get process stdout to read server log
	Stderr() (io.ReadCloser, error) // Get process stderr to read server log
	Stdin() (io.WriteCloser, error) // Get process stdin to write to server
	Write(p []byte) (int, error)    // Write to stdin
	Wait() (int64, error)           // Wait server process end return exit code, if exit then 0 return error
	Tar(w io.Writer) error          // Create backup in tar file
	Zip(w io.Writer) error          // Create backup in zip file
	Fs() (fs.FS, error)             // Get server fs.FS
}
