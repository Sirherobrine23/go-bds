package overlayfs

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

type FileTestStruct struct {
	Name, SaveOn string      // File name and Save on
	Mode         fs.FileMode // File permission
	Content      []byte      // File body
	Write        bool        // This file write in test
}

func initLayers(tmp string, NodesFiles []FileTestStruct) ([]string, error) {
	nodesLOW := []string{}
	for _, file := range NodesFiles {
		if file.Write {
			continue // Skip Write files
		}

		dir := filepath.Join(tmp, file.SaveOn)
		if !slices.Contains(nodesLOW, dir) {
			nodesLOW = append(nodesLOW, dir)
		}

		folderDir := filepath.Join(dir, filepath.Dir(file.Name))
		if _, err := os.Stat(folderDir); err != nil && os.IsNotExist(err) {
			if err := os.MkdirAll(folderDir, file.Mode.Perm()); err != nil {
				return nil, err
			}
		}

		err := os.WriteFile(filepath.Join(dir, file.Name), file.Content, file.Mode.Perm())
		if err != nil {
			return nil, err
		}
	}
	return nodesLOW, nil
}

func roAssert(overlay *Overlayfs, file FileTestStruct) error {
	if file.Content == nil {
		stat, err := overlay.Stat(file.Name)
		if err != nil {
			return err
		} else if stat.Mode().Perm() != file.Mode.Perm() {
			return fmt.Errorf("invalid permision writed file, %d != %d", file.Mode.Perm(), stat.Mode().Perm())
		}
		return nil
	}

	data, err := overlay.ReadFile(file.Name)
	if err != nil {
		return err
	} else if !bytes.Equal(data, file.Content) {
		return fmt.Errorf("invalid file content, %v != %v", file.Content, data)
	}
	return nil
}

func rwAssert(overlay *Overlayfs, file FileTestStruct) error {
	if !file.Write {
		return roAssert(overlay, file)
	} else if _, err := overlay.Stat(file.Name); err == nil {
		return fmt.Errorf("invalid write File, file exist (%q)", file.Name)
	}

	if err := overlay.MkdirAll(filepath.Dir(file.Name), file.Mode.Perm()); err != nil {
		return errors.Join(err, fmt.Errorf("cannot create dir (%q)", filepath.Dir(file.Name)))
	} else if err = overlay.WriteFile(file.Name, file.Content, file.Mode.Perm()); err != nil {
		return errors.Join(err, fmt.Errorf("cannot write file (%q)", file.Name))
	}

	{
		ossPath := filepath.Join(overlay.Upper, file.Name)
		if file.Content == nil {
			stat, err := os.Stat(ossPath)
			if err != nil {
				return err
			} else if stat.Mode().Perm() != file.Mode.Perm() {
				return fmt.Errorf("invalid permision in writed file, %d != %d", file.Mode.Perm(), stat.Mode().Perm())
			}
			return nil
		}

		data, err := os.ReadFile(ossPath)
		if err != nil {
			return err
		} else if !bytes.Equal(data, file.Content) {
			return fmt.Errorf("invalid file content, %v != %v", file.Content, data)
		}
	}

	{
		data, err := overlay.ReadFile(file.Name)
		if err != nil {
			return err
		} else if !bytes.Equal(data, file.Content) {
			return fmt.Errorf("invalid file content, %v != %v", file.Content, data)
		}
	}

	return nil
}

func TestMergefsRO(t *testing.T) {
	tmp := t.TempDir()
	nodesLOW, err := initLayers(tmp, ROFiles)
	if err != nil {
		t.Skip(err.Error())
		return
	}

	target := filepath.Join(tmp, "target")
	if err := os.Mkdir(target, 0666); err != nil {
		t.Skip(err.Error())
		return
	}

	overlayTest := &Overlayfs{
		Target: target,
		Lower:  nodesLOW,
	}

	for _, file := range ROFiles {
		if err = roAssert(overlayTest, file); err != nil {
			t.Error(err)
			return
		}
	}
}

func TestMergefsRW(t *testing.T) {
	tmp := t.TempDir()
	nodesLOW, err := initLayers(tmp, RWFiles)
	if err != nil {
		t.Skip(err.Error())
		return
	}
	rootUpper := filepath.Join(tmp, "upper")
	_ = os.Mkdir(rootUpper, 0755)

	overlayTest := &Overlayfs{
		Lower: nodesLOW,
		Upper: rootUpper,
	}

	for _, file := range RWFiles {
		if err = rwAssert(overlayTest, file); err != nil {
			t.Error(err)
			return
		}
	}

	targetTest := RWFiles[0]
	if err = overlayTest.Symlink(targetTest.Name, "testlink.txt"); err != nil {
		t.Error(errors.Join(errors.New("cannot create syslink"), err))
		return
	}

	data, err := overlayTest.ReadFile("testlink.txt")
	if err != nil {
		t.Error(errors.Join(errors.New("cannot read link"), err))
		return
	} else if !bytes.Equal(data, targetTest.Content) {
		t.Errorf("link not same, datas: %v != %v", data, targetTest.Content)
		return
	}

	if err := overlayTest.Remove(targetTest.Name); err != nil {
		t.Error(errors.Join(errors.New("cannot delete file"), err))
		return
	}

	entrys, err := overlayTest.ReadDir(".")
	if err != nil {
		t.Error(errors.Join(errors.New("cannot list files in root"), err))
		return
	}
	entryNames := []string{}
	for _, k := range entrys {
		entryNames = append(entryNames, fmt.Sprintf("%s: isSyslink %v", k.Name(), k.Type() == fs.ModeSymlink))
	}
	if slices.Contains(entryNames, targetTest.Name) {
		t.Error("invalid list dir")
		return
	}

	perm := fs.FileMode(0777)
	targetTest = RWFiles[1]
	err = overlayTest.Chmod(targetTest.Name, perm)
	if err != nil {
		t.Error(errors.Join(errors.New("cannot change chmod"), err))
		return
	}

	stat, err := overlayTest.Stat(targetTest.Name)
	if err != nil {
		t.Error(errors.Join(errors.New("cannot get chmod from change mod"), err))
		return
	} else if stat.Mode().Perm() != perm {
		t.Error(errors.Join(fmt.Errorf("permission is mismatch %s != %s", perm, stat.Mode().Perm()), err))
		return
	}
}

// Files config to test
var (
	ROFiles = []FileTestStruct{
		{
			Name:    "top1.txt",
			SaveOn:  "low3",
			Mode:    0755,
			Content: []byte("Hello world"),
		},
		{
			Name:    "top2.txt",
			SaveOn:  "low1",
			Mode:    0755,
			Content: []byte{0, 2, 7, 9, 244, 76, 81, 01, 05, 0003},
		},
		{
			Name:    "low/top3.txt",
			SaveOn:  "low3",
			Mode:    0755,
			Content: nil,
		},
		{
			Name:    "Google/com/br/search/golang.txt",
			SaveOn:  "low3",
			Mode:    0755,
			Content: []byte("Go é uma linguagem de programação criada pela Google e lançada em código livre em novembro de 2009. É uma linguagem compilada e focada em produtividade e programação concorrente,[6] baseada em trabalhos feitos no sistema operacional chamado Inferno.[7] O projeto inicial da linguagem foi feito em setembro de 2007 por Robert Griesemer, Rob Pike e Ken Thompson.[6] Atualmente, há implementações para Windows, Linux, Mac OS X e FreeBSD.[4]"), // GO is a programming language created by Google and released in Free Code in November 2009. It is a compiled language focused on productivity and competing programming, [6] based on works done on the operating system called hell. [7] The initial language project was made in September 2007 by Robert Griesemer, Rob Pike and Ken Thompson. [6] Currently, there are implementations for Windows, Linux, Mac OS X and FreeBSD. [4]
		},
	}
	RWFiles = []FileTestStruct{
		{
			Name:    "toprd1.txt",
			SaveOn:  "low1",
			Mode:    0755,
			Content: []byte{0, 0, 0, 1, 1, 1, 0, 2, 8},
		},
		{
			Name:    "toprw2.txt",
			Mode:    0755,
			Write:   true,
			Content: []byte{5, 8, 23, 80, 40, 21},
		},
	}
)
