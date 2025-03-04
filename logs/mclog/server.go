package mclog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"path"
	"runtime"
	"slices"
	"strings"

	"sirherobrine23.com.br/go-bds/go-bds/logs"
)

func writeJson(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	jsenc := json.NewEncoder(w)
	jsenc.SetIndent("", "  ")
	jsenc.Encode(v)
}

func NewHandler(limits Limits, fss FileSystem) http.Handler {
	fss.Mkdir("v2", 0755) // Ignore error's
	mux, v1, v2 := http.NewServeMux(), http.NewServeMux(), http.NewServeMux()

	// mclogs limits
	v1Limits, _ := json.MarshalIndent(map[string]any{
		"storageTime": int(limits.StorageTime.Seconds()),
		"maxLength":   limits.MaxLength,
		"maxLines":    limits.MaxLines,
	}, "", "  ")

	// V1 from mclogs
	v1.HandleFunc("GET /limits", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(v1Limits)
	})

	// Storage log
	v1.HandleFunc("POST /log", func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(r.Header.Values("Content-Type"), "application/x-www-form-urlencoded") {
			r.ParseForm()
		}
		if !(r.Form.Has("content") || r.Form.Get("content") == "") {
			writeJson(w, 400, map[string]any{
				"success": false,
				"error":   "Required POST argument 'content' is empty.",
			})
			return
		}

		if limits.MaxLines > 0 || limits.MaxLength > 0 {
			lines := strings.Split(r.Form.Get("content"), "\n")
			content := strings.Join(lines[:min(len(lines)-1, int(limits.MaxLines))], "\n")
			if limits.MaxLength > 0 {
				content = content[:min(len(content), int(limits.MaxLength))]
			}
			r.Form.Set("content", content)
		}

		if _, err := logs.ParseString(r.Form.Get("content")); err != nil {
			writeJson(w, 400, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		id, file, err := CreateID(fss, ".")
		if err != nil {
			writeJson(w, 500, map[string]any{
				"success": false,
				"error":   fmt.Errorf("canont create file and id: %s", err).Error(),
			})
			return
		}
		defer file.Close()

		if _, err = io.WriteString(file, r.Form.Get("content")); err != nil {
			writeJson(w, 500, map[string]any{
				"success": false,
				"error":   fmt.Errorf("canont write log to file: %s", err).Error(),
			})
			return
		}
		writeJson(w, 200, map[string]any{
			"success": true,
			"id":      id,
		})
	})

	// Get log
	v1.HandleFunc("GET /raw/{id}", func(w http.ResponseWriter, r *http.Request) {
		file, err := fss.Open(r.PathValue("id"))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				writeJson(w, 404, map[string]any{
					"success": false,
					"error":   "Log not found.",
				})
			} else {
				writeJson(w, 400, map[string]any{
					"success": false,
					"error":   err.Error(),
				})
			}
			return
		}
		defer file.Close()
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)
		io.Copy(w, file)
	})

	// Get log insights
	v1.HandleFunc("GET /insights/{id}", func(w http.ResponseWriter, r *http.Request) {
		file, err := fss.Open(r.PathValue("id"))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				writeJson(w, 404, map[string]any{
					"success": false,
					"error":   "Log not found.",
				})
			} else {
				writeJson(w, 400, map[string]any{
					"success": false,
					"error":   err.Error(),
				})
			}
			return
		}
		defer file.Close()
		buff, err := io.ReadAll(file)
		if err != nil {
			writeJson(w, 500, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		log, err := logs.ParseBuffer(buff)
		if err != nil {
			writeJson(w, 500, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		writeJson(w, 500, ConvertLogs(r.PathValue("id"), log))
	})

	// Analyse a log without saving it
	v1.HandleFunc("POST /analyse", func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(r.Header.Values("Content-Type"), "application/x-www-form-urlencoded") {
			r.ParseForm()
		}
		if !(r.Form.Has("content") || r.Form.Get("content") == "") {
			writeJson(w, 400, map[string]any{
				"success": false,
				"error":   "Required POST argument 'content' is empty.",
			})
			return
		}

		if limits.MaxLines > 0 || limits.MaxLength > 0 {
			lines := strings.Split(r.Form.Get("content"), "\n")
			content := strings.Join(lines[:min(len(lines)-1, int(limits.MaxLines))], "\n")
			if limits.MaxLength > 0 {
				content = content[:min(len(content), int(limits.MaxLength))]
			}
			r.Form.Set("content", content)
		}

		log, err := logs.ParseString(r.Form.Get("content"))
		if err != nil {
			writeJson(w, 400, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		writeJson(w, 200, ConvertLogs("", log))
	})

	// Server v2
	v2Info, _ := json.MarshalIndent(map[string]any{"runtime": runtime.Version(), "limits": limits}, "", "  ")
	v2.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(v2Info)
		w.Write(nil)
	})

	v2.HandleFunc("POST /{$}", func(w http.ResponseWriter, r *http.Request) {
		logString, parsedLogs := []string{}, map[string]logs.Log{}
		if slices.Contains(r.Header.Values("Content-Type"), "application/octet-stream") {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("cannot read json body: %s", err).Error()})
				return
			}
			logString = append(logString, string(body))
		} else if slices.Contains(r.Header.Values("Content-Type"), "multipart/form-data") {
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("cannot parse form: %s", err).Error()})
				return
			}
			for _, content := range r.MultipartForm.Value {
				logString = append(logString, content...)
			}
			for _, content := range r.MultipartForm.File {
				for _, file := range content {
					info, err := file.Open()
					if err != nil {
						writeJson(w, 400, map[string]any{"error": fmt.Errorf("cannot read json body: %s", err).Error()})
						return
					}
					defer info.Close()
					body, err := io.ReadAll(info)
					if err != nil {
						writeJson(w, 400, map[string]any{"error": fmt.Errorf("cannot read json body: %s", err).Error()})
						return
					}
					logString = append(logString, string(body))
				}
			}
		} else if slices.Contains(r.Header.Values("Content-Type"), "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err != nil {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("cannot parse form: %s", err).Error()})
				return
			}
			for _, content := range r.Form {
				logString = append(logString, content...)
			}
		} else if slices.Contains(r.Header.Values("Content-Type"), "application/json") {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("cannot read json body: %s", err).Error()})
				return
			}

			var data any
			if err = json.Unmarshal(body, &data); err != nil {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("cannot parse json: %s", err).Error()})
				return
			}

			switch v := data.(type) {
			case []string:
				logString = v
			case map[string]string:
				logString = slices.Collect(maps.Values(v))
			default:
				writeJson(w, 400, map[string]any{"error": "invalid data type", "tips": "[]string, map[string]string"})
				return
			}
		} else {
			writeJson(w, 400, map[string]any{"error": "require Content-Type in header"})
			return
		}

		err := error(nil)
		for _, log := range logString {
			log = strings.TrimSpace(log)
			sum := sha256.Sum256([]byte(log))
			if parsedLogs[hex.EncodeToString(sum[:])], err = logs.ParseString(log); err != nil {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("cannot parse json: %s", err).Error()})
				return
			}
		}

		writeJson(w, 200, parsedLogs)
	})

	v2.HandleFunc("POST /raw", func(w http.ResponseWriter, r *http.Request) {
		if !slices.Contains(r.Header.Values("Content-Type"), "application/octet-stream") {
			writeJson(w, 400, map[string]any{"error": "require application/octet-stream on Content-Type header"})
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJson(w, 400, map[string]any{"error": fmt.Errorf("cannot get full body: %s", err).Error()})
			return
		}

		body = []byte(strings.TrimSpace(string(body)))
		if _, err = logs.ParseBuffer(body); err != nil {
			writeJson(w, 400, map[string]any{"error": fmt.Errorf("invalid log: %s", err).Error()})
			return
		}

		fileSum := sha256.Sum256(body)
		fileID := hex.EncodeToString(fileSum[:])
		remoteFile, err := fss.Open(path.Join("v2", fileID))
		if remoteFile != nil {
			remoteFile.Close() // Close file
		}

		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("fs error: %s", err).Error()})
				return
			}
			create, err := fss.Create(path.Join("v2", fileID))
			if err != nil {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("fs error: %s", err).Error()})
				return
			}
			create.Write(body)
			create.Close()
		}
		writeJson(w, 200, map[string]any{"id": fileID})
	})

	v2.HandleFunc("GET /raw/{id}", func(w http.ResponseWriter, r *http.Request) {
		file, err := fss.Open(path.Join("v2", r.PathValue("id")))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				writeJson(w, 404, map[string]any{"error": "file not exists"})
			} else {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("fs error: %s", err).Error()})
			}
			return
		}
		defer file.Close()
		io.Copy(w, file)
	})
	v2.HandleFunc("GET /insight/{id}", func(w http.ResponseWriter, r *http.Request) {
		file, err := fss.Open(path.Join("v2", r.PathValue("id")))
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				writeJson(w, 404, map[string]any{"error": "file not exists"})
			} else {
				writeJson(w, 400, map[string]any{"error": fmt.Errorf("fs error: %s", err).Error()})
			}
			return
		}
		defer file.Close()
		body, err := io.ReadAll(file)
		if err != nil {
			writeJson(w, 400, map[string]any{"error": fmt.Errorf("read: %s", err).Error()})
			return
		}
		log, err := logs.ParseBuffer(body)
		if err != nil {
			writeJson(w, 400, map[string]any{"error": fmt.Errorf("logs parse: %s", err).Error()})
			return
		}
		writeJson(w, 200, log)
	})

	mux.Handle("/1/", http.StripPrefix("/1", v1))
	mux.Handle("/v1/", http.StripPrefix("/v1", v1))
	mux.Handle("/v2/", http.StripPrefix("/v2", v2))
	return mux
}
