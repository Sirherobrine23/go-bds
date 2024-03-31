package zip_tar

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/zhyee/zipstream"
)

func Test() {
	resp, err := http.Get("https://github.com/golang/go/archive/refs/tags/go1.16.10.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	zr := zipstream.NewReader(resp.Body)

	for {
		e, err := zr.GetNextEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("unable to get next entry: %s", err)
		}

		log.Println("entry name: ", e.Name)
		log.Println("entry comment: ", e.Comment)
		log.Println("entry reader version: ", e.ReaderVersion)
		log.Println("entry modify time: ", e.Modified)
		log.Println("entry compressed size: ", e.CompressedSize64)
		log.Println("entry uncompressed size: ", e.UncompressedSize64)
		log.Println("entry is a dir: ", e.IsDir())

		if !e.IsDir() {
			rc, err := e.Open()
			if err != nil {
				log.Fatalf("unable to open zip file: %s", err)
			}
			content, err := io.ReadAll(rc)
			if err != nil {
				log.Fatalf("read zip file content fail: %s", err)
			}

			log.Println("file length:", len(content))

			if uint64(len(content)) != e.UncompressedSize64 {
				log.Fatalf("read zip file length not equal with UncompressedSize64")
			}
			if err := rc.Close(); err != nil {
				log.Fatalf("close zip entry reader fail: %s", err)
			}
		}
	}
}

func ZipToTar(in io.Reader, out io.Writer) error {
	zr := zipstream.NewReader(in)
	tarStr := tar.NewWriter(out)
	
	for {
		fileEntry, err := zr.GetNextEntry()
		if err == io.EOF {
			err = tarStr.Close()
			if err != nil {
				return fmt.Errorf("unable close tar: %s", err)
			}
			break
		}

		if err != nil {
			return fmt.Errorf("unable to get next entry: %s", err)
		}

		fileInfo := fileEntry.FileInfo()
		head := tar.Header{
			Name: fileEntry.Name,
			Size: fileInfo.Size(),
			ChangeTime: fileEntry.Modified,
		}
		tarStr.WriteHeader(&head)
		if !fileEntry.IsDir() {
			fileContent, err := fileEntry.Open()
			if err != nil {
				return fmt.Errorf("unable to get entry content: %s", err)
			}

			_, err = io.Copy(tarStr, fileContent)
			if err != nil {
				return err
			}

			err = fileContent.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}