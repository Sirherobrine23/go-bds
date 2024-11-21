package descompress

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/ulikunitz/xz"
)

var BodyTest = []byte("Test body and content body")

func TestGzip(t *testing.T) {
	compressedBody := new(bytes.Buffer)

	gz := gzip.NewWriter(compressedBody)
	defer gz.Close()
	if _, err := gz.Write(BodyTest); err != nil {
		t.Error(err)
		return
	}
	gz.Close()

	ungz, err := NewDescompress(compressedBody)
	if err != nil {
		t.Error(err)
		return
	}
	defer ungz.(io.Closer).Close()

	readBuff := make([]byte, len(BodyTest))
	if _, err := ungz.Read(readBuff); err != nil && err != io.EOF {
		t.Error(err)
		return
	}
	ungz.(io.Closer).Close()

	if !bytes.Equal(readBuff, BodyTest) {
		t.Errorf("Content not is sames")
	}
}

func TestXZ(t *testing.T) {
	compressedBody := new(bytes.Buffer)

	xzCompressed, err := xz.NewWriter(compressedBody)
	if err != nil {
		t.Error(err)
		return
	}
	defer xzCompressed.Close()

	if _, err := xzCompressed.Write(BodyTest); err != nil {
		t.Error(err)
		return
	}
	xzCompressed.Close()

	xzDescompressed, err := NewDescompress(compressedBody)
	if err != nil {
		t.Error(err)
		return
	}

	readBuff := make([]byte, len(BodyTest))
	if _, err := xzDescompressed.Read(readBuff); err != nil && err != io.EOF {
		t.Error(err)
		return
	} else if !bytes.Equal(readBuff, BodyTest) {
		t.Errorf("Content not is sames")
	}
}
func TestZstd(t *testing.T) {
	compressedBody := new(bytes.Buffer)

	zstdCompressed, err := xz.NewWriter(compressedBody)
	if err != nil {
		t.Error(err)
		return
	}
	defer zstdCompressed.Close()

	if _, err := zstdCompressed.Write(BodyTest); err != nil {
		t.Error(err)
		return
	}
	zstdCompressed.Close()

	zstdDescompressed, err := NewDescompress(compressedBody)
	if err != nil {
		t.Error(err)
		return
	}

	readBuff := make([]byte, len(BodyTest))
	if _, err := zstdDescompressed.Read(readBuff); err != nil && err != io.EOF {
		t.Error(err)
		return
	} else if !bytes.Equal(readBuff, BodyTest) {
		t.Errorf("Content not is sames")
	}
}
