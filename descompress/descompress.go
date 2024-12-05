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
func NewDescompress(r io.Reader) (io.Reader, error) {
	buf := make([]byte, 15)
	_, err := r.Read(buf)
	if err != nil {
		return r, err
	}

	r = io.MultiReader(bytes.NewReader(buf), r)

	switch {
	case bytes.HasPrefix(buf, []byte{0x1F, 0x8B, 0x08}):
		return gzip.NewReader(r)
	case bytes.HasPrefix(buf, []byte{0x5A, 0x42, buf[2], 0x68, 0x41, 0x31, 0x26, 0x59, 0x59, 0x53}):
		return bzip2.NewReader(r), nil
	case bytes.HasPrefix(buf, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}):
		return xz.NewReader(r)
	case bytes.HasPrefix(buf, []byte{0xB5, 0x28, 0xFD, 0x2F}):
		return zstd.NewReader(r)
	case ((buf[0] == 0x78) && (buf[1] == 1 || buf[1] == 0x9C || buf[1] == 0xDA)):
		return flate.NewReader(r), nil
	default:
		return r, nil
	}
}
