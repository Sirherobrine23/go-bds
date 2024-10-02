package mergefs

import (
	"bytes"
	crypto "crypto/rand"
	"encoding/hex"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
)

func TestMergefs(t *testing.T) {
	tmpFolder, err := os.MkdirTemp(os.TempDir(), "testMergefs_*")
	if err != nil {
		t.Skipf("Cannot make temporary folder to test Mergefs, error: %s", err)
		t.SkipNow()
		return
	}
	t.Logf("Temp folder %q", tmpFolder)
	defer os.RemoveAll(tmpFolder)

	var mergeFolders Mergefs
	mergeFolders.TopLayer = filepath.Join(tmpFolder, "top")
	mergeFolders.LowerLayers = []string{
		filepath.Join(tmpFolder, "low1"),
		filepath.Join(tmpFolder, "low2"),
		filepath.Join(tmpFolder, "gotFolder"),
	}
	for _, f := range append(mergeFolders.LowerLayers, mergeFolders.TopLayer) {
		os.MkdirAll(f, 0666)
	}

	type locations struct {
		Root, File string
		Content    []byte
	}
	files := []locations{}
	for range rand.IntN(128) {
		rootLow := mergeFolders.LowerLayers[rand.IntN(len(mergeFolders.LowerLayers))]
		tempFile, err := os.CreateTemp(rootLow, "*.txt")
		if err != nil {
			continue
		}
		content := make([]byte, 256)
		crypto.Read(content)
		files = append(files, locations{
			File:    filepath.Base(tempFile.Name()),
			Root:    rootLow,
			Content: content,
		})
		os.WriteFile(tempFile.Name(), content, 0666)
	}

	t.Run("Stat", func(t *testing.T) {
		_, err = mergeFolders.Stat("")
		if !os.IsNotExist(err) {
			t.Error(err)
			return
		}
		for range len(files) / 2 {
			randFile := files[rand.IntN(len(files))]
			t.Run(randFile.File, func(t *testing.T) {
				_, err = mergeFolders.Stat(randFile.File)
				if err != nil {
					t.Error(err)
					return
				}
			})
		}
	})

	t.Run("Read", func(t *testing.T) {
		for range len(files) / 2 {
			randFile := files[rand.IntN(len(files))]
			data, err := mergeFolders.ReadFile(randFile.File)
			if err != nil {
				t.Error(err)
				return
			} else if !bytes.Equal(data, randFile.Content) {
				t.Errorf("Bytes not is equal, Required %q, From Mergefs %q", hex.EncodeToString(randFile.Content), hex.EncodeToString(data))
				return
			}
		}
	})
}
