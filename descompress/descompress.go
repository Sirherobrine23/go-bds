// Detect stream compression and return new stream with file descompressed, if not detect compression return initial stream
package descompress

import (
	"bytes"
	"compress/bzip2"
	"compress/flate"
	"compress/gzip"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

// Auto descompress stream
func NewDescompress(inputReader io.Reader) (io.ReadCloser, error) {
	// Read fist bytes to detect file compression
	buf := make([]byte, 15)
	if _, err := inputReader.Read(buf); err != nil {
		if v, ok := inputReader.(io.ReadCloser); ok {
			return v, err
		}
		return io.NopCloser(inputReader), err
	}

	// Concat read bytes
	inputReader = io.MultiReader(bytes.NewReader(buf), inputReader)

	switch {
	// Gzip
	case bytes.HasPrefix(buf, []byte{0x1F, 0x8B, 0x08}):
		return gzip.NewReader(inputReader)

	// BZip
	case bytes.HasPrefix(buf, []byte{0x5A, 0x42, buf[2], 0x68, 0x41, 0x31, 0x26, 0x59, 0x59, 0x53}):
		return io.NopCloser(bzip2.NewReader(inputReader)), nil

	// XZ
	case bytes.HasPrefix(buf, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}):
		read, err := xz.NewReader(inputReader)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(read), nil

	// Zstd
	case bytes.HasPrefix(buf, []byte{0xB5, 0x28, 0xFD, 0x2F}):
		read, err := zstd.NewReader(inputReader)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(read), nil

	// Deflate
	case ((buf[0] == 0x78) && (buf[1] == 1 || buf[1] == 0x9C || buf[1] == 0xDA)):
		return flate.NewReader(inputReader), nil

	// Input stream
	default:
		return io.NopCloser(inputReader), nil
	}
}
