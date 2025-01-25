package overlayfs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGoMergefs(t *testing.T) {
	root, err := os.MkdirTemp(t.TempDir(), "mergefs*")
	if err != nil {
		t.Skipf("Cannot make root folder to test go-mergefs, error: %s", err.Error())
		return
	}

	t.Run("Read-Only", func(t *testing.T) {
		Mergefs := NewOverlayFS("", "", "", filepath.Join(root, "ronly/low1"), filepath.Join(root, "ronly/low2"))
		for _, mk := range Mergefs.Lower {
			if err := os.MkdirAll(mk, 0777); err != nil {
				t.Skipf("cannot create folder %q, skiped because cannot make folders, error: %s", mk, err.Error())
				return
			}
		}

		contentBody, FilesList := []byte("Golang is best"), []string{
			".",                             // Tentar pegar a camda Upper
			"file1.txt",                     // Arquivo na Root
			"file2.txt",                     // Secondo aquivo na Root
			"long/long/file3.txt",           // Terceiro arquivo, mas com caminho longo
			"long/long/longshort/file4.txt", // Quarto arquivo, mas com caminho longo
		}

		for fileIndex := range FilesList {
			if fileIndex == 0 {
				continue
			}
			rootPath := Mergefs.Lower[0]
			if (fileIndex % 2) == 0 {
				rootPath = Mergefs.Lower[1]
			}

			// Create dir if not exists
			if dir := filepath.Dir(FilesList[fileIndex]); dir != "" {
				if _, err := os.Stat(filepath.Join(rootPath, dir)); err != nil {
					if err = os.MkdirAll(filepath.Join(rootPath, dir), 0777); err != nil {
						t.Error("cannot make files or folder to test")
						return
					}
				}
			}

			if err := os.WriteFile(filepath.Join(rootPath, FilesList[fileIndex]), contentBody, 0777); err != nil {
				t.Error("cannot make files or folder to test", err)
				return
			}
		}

		for _, filePath := range FilesList {
			// Stat
			t.Logf("Stat=%q", filePath)
			_, err := Mergefs.Stat(filePath)
			if filePath == FilesList[0] && err != nil {
				if !os.IsPermission(err) {
					t.Error("Stat required return Bad Permission to read-only")
					return
				}
			} else if err != nil {
				t.Error(err)
				return
			}

			if filePath != FilesList[0] {
				t.Logf("Read=%q", filePath)
				content, err := Mergefs.ReadFile(filePath)
				if err != nil {
					t.Error(err)
					return
				} else if !bytes.Equal(content, contentBody) {
					t.Error("The file body are not the same")
					return
				}
			}
		}
	})

	t.Run("Read-Write", func(t *testing.T) {
		Mergefs := NewOverlayFS("", filepath.Join(root, "rw/up"), "", filepath.Join(root, "rw/low1"), filepath.Join(root, "rw/low2"))
		for _, mk := range append(Mergefs.Lower, Mergefs.Upper) {
			if err := os.MkdirAll(mk, 0777); err != nil {
				t.Skipf("cannot create folder %q, skiped because cannot make folders, error: %s", mk, err.Error())
				return
			}
		}

		type RW struct {
			path   string
			P1, P2 []byte
		}

		writeFile := []RW{
			{"fileLong/file1.txt", []byte("golang made"), []byte("designed by Google")},
			{"fileLong/longshort/file2.txt", []byte("designed by Google"), []byte("google.com")},
			{"root3.txt", []byte("youtube.com"), []byte("hangouts.google.com")},
			{"short/file4.txt", []byte("designed by Google"), []byte("google.com")},
			{"world/wide/web/file5.txt", []byte("designed by Google"), []byte("google.com")},
			{"search/world/wide/web/file6.txt", []byte("designed by Google"), []byte("google.com")},
			{"search/world/wide/web/file7.txt", []byte("designed by Google"), []byte("google.com")},
			{"minecraft/java/bedrock/global/maneger8.txt", []byte("designed by Google"), []byte("google.com")},
		}

		for _, fileToMake := range writeFile {
			{
				t.Logf("Testing %q P1", fileToMake.path)
				if dir := filepath.Dir(fileToMake.path); dir != "" {
					if _, err := Mergefs.Stat(fileToMake.path); os.IsNotExist(err) {
						if err = Mergefs.MkdirAll(dir, 0777); err != nil {
							t.Error(err)
							return
						}
					}
				}

				if err := Mergefs.WriteFile(fileToMake.path, fileToMake.P1, 0777); err != nil {
					t.Error(err)
					return
				}
				content, err := Mergefs.ReadFile(fileToMake.path)
				if err != nil {
					t.Error(err)
					return
				} else if !bytes.Equal(content, fileToMake.P1) {
					t.Error("Body mismatch")
					return
				}
			}
			{
				t.Logf("Testing %q P2", fileToMake.path)
				if dir := filepath.Dir(fileToMake.path); dir != "" {
					if _, err := Mergefs.Stat(fileToMake.path); os.IsNotExist(err) {
						if err = Mergefs.MkdirAll(dir, 0777); err != nil {
							t.Error(err)
							return
						}
					}
				}

				if err := Mergefs.WriteFile(fileToMake.path, fileToMake.P2, 0777); err != nil {
					t.Error(err)
					return
				}

				content, err := Mergefs.ReadFile(fileToMake.path)
				if err != nil {
					t.Error(err)
					return
				} else if !bytes.Equal(content, fileToMake.P2) {
					t.Error("Body mismatch")
					return
				}
			}
		}
	})
}
