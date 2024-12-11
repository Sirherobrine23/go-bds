package mclog

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Storage interface {
	fs.FS
	Save(name string, body io.Reader) (int64, error) // Save file in storage
	Remove(name string) error                        // Remove file from storage
}

func NewHandler(serverLimits Limits, storage Storage) *http.ServeMux {
	control := http.NewServeMux()

	control.HandleFunc("GET /1/limits", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(Limits{
			StorageTime: serverLimits.StorageTime / time.Second,
			MaxLength:   serverLimits.MaxLength,
			MaxLines:    serverLimits.MaxLines,
		})
	})

	control.HandleFunc("GET /1/raw/{id}", func(w http.ResponseWriter, r *http.Request) {
		logId := r.PathValue("id")
		body, err := storage.Open(logId)
		if err != nil && errors.Is(err, fs.ErrNotExist) {
			w.WriteHeader(404)
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			enc.Encode(MclogResponseStatus{
				Success:      false,
				ErrorMessage: "Log file not exists",
			})
			return // Close function
		}
		defer body.Close()

		w.Header().Set("Content-Type", "application/octet-stream") // Set file stream
		if stat, err := body.Stat(); err == nil {
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("content-length", strconv.FormatInt(stat.Size(), 10))
		}

		w.WriteHeader(200)
		go io.Copy(w, body)
	})

	control.HandleFunc("GET /1/insights/{id}", func(w http.ResponseWriter, r *http.Request) {
		logId := r.PathValue("id") // get log id

		// Open log file
		logBody, err := storage.Open(logId)
		if err != nil {
			// Return if log not exists
			if errors.Is(err, fs.ErrNotExist) {
				w.WriteHeader(404)
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				enc.Encode(MclogResponseStatus{
					Success:      false,
					ErrorMessage: "log not exists",
				})
				return // close function
			}

			// Return sys error
			w.WriteHeader(500)
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			enc.Encode(MclogResponseStatus{
				Success:      false,
				ErrorMessage: err.Error(),
			})
			return // close function
		}
		defer logBody.Close()

		// Parse log body
		var logInsight Insights
		if err := logInsight.ParseLogFile(logBody); err != nil {
			w.WriteHeader(500)
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			enc.Encode(MclogResponseStatus{
				Success:      false,
				ErrorMessage: err.Error(),
			})
			return
		}

		// return log insight
		w.WriteHeader(200)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(logInsight)
	})

	control.HandleFunc("POST /1/log", func(w http.ResponseWriter, r *http.Request) {
		var randomId string
		for randomId == "" {
			buff := make([]byte, 8)
			if _, err := rand.Read(buff); err != nil && !errors.Is(err, io.EOF) {
				w.WriteHeader(500)
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				enc.Encode(MclogResponseStatus{
					Success:      false,
					ErrorMessage: "cannot make log id",
				})
				return
			}
			randomId = hex.EncodeToString(buff)
			if f, err := storage.Open(randomId); err == nil {
				f.Close()
				randomId = ""
			}
		}

		if slices.Contains(r.Header.Values("Content-Type"), "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err != nil {
				w.WriteHeader(500)
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				enc.Encode(MclogResponseStatus{
					Success:      false,
					ErrorMessage: err.Error(),
				})
				return
			}

			// Check if body have "content"
			if !r.Form.Has("content") {
				w.WriteHeader(400)
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				enc.Encode(MclogResponseStatus{
					Success:      false,
					ErrorMessage: "require the 'content' in body",
				})
				return
			}

			// Attemp write log file
			if _, err := storage.Save(randomId, strings.NewReader(r.Form.Get("content"))); err != nil {
				w.WriteHeader(500)
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				enc.Encode(MclogResponseStatus{
					Success:      false,
					ErrorMessage: err.Error(),
				})
				return
			}
		} else {
			// Close body after run function
			defer r.Body.Close()

			logStream := io.Reader(r.Body)
			if serverLimits.MaxLength > 0 {
				logStream = io.LimitReader(r.Body, serverLimits.MaxLength)
			}

			// Attemp write log file
			if _, err := storage.Save(randomId, logStream); err != nil {
				w.WriteHeader(500)
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				enc.Encode(MclogResponseStatus{
					Success:      false,
					ErrorMessage: err.Error(),
				})
				return
			}
		}

		// Write success log id
		w.WriteHeader(201)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(MclogResponseStatus{
			Success: true,
			Id:      randomId,
		})
	})

	return control
}
