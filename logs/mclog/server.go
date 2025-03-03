package mclog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
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

	mux.Handle("/1/", http.StripPrefix("/1", v1))
	mux.Handle("/v2/", http.StripPrefix("/v2", v2))
	return mux
}
