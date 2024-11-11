package descompress

import (
	"compress/bzip2"
	"compress/flate"
	"compress/gzip"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

type seekReader struct {
	r    io.Reader
	buff []byte
}

func (rs *seekReader) Read(p []byte) (int, error) {
	if len(rs.buff) > 0 {
		n := copy(p, rs.buff)
		rs.buff = rs.buff[n:]
		return n, nil
	}
	return rs.r.Read(p)
}

// Auto descompress stream
func NewDescompress(r io.Reader) (io.Reader, error) {
	buf := make([]byte, 15)
	_, err := r.Read(buf)
	if err != nil {
		return r, err
	}

	switch {
	case (buf[0] == 0xFD) && (buf[1] == 0x37) && (buf[2] == 0x7A) && (buf[3] == 0x58) && (buf[4] == 0x5A) && (buf[5] == 0x00):
		return xz.NewReader(&seekReader{r, buf})
	case (buf[0] == 0x1F && (buf[1] == 0x8B) && (buf[2] == 0x08)):
		return gzip.NewReader(&seekReader{r, buf})
	case ((buf[0] == 0x5A) && (buf[1] == 0x42) && (buf[3] == 0x68) && (buf[4] == 0x41) && (buf[5] == 0x31) && (buf[6] == 0x26) && (buf[7] == 0x59) && (buf[8] == 0x59) && (buf[9] == 0x53)):
		return bzip2.NewReader(&seekReader{r, buf}), nil
	case ((buf[0] == 0x78) && (buf[1] == 1 || buf[1] == 0x9c || buf[1] == 0xda)):
		return flate.NewReader(&seekReader{r, buf}), nil
	case ((buf[0] == 0xB5) && (buf[1] == 0x28) && (buf[2] == 0xFD) && (buf[3] == 0x2F)):
		return zstd.NewReader(&seekReader{r, buf})
	default:
		return r, nil
	}
}
