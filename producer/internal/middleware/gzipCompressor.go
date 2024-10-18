package middleware

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

type GzipWriter struct {
	ResWriter       http.ResponseWriter
	Writer          *gzip.Writer
	GzipWriterMutex sync.Mutex
	GzipFlag        bool
}

func (w *GzipWriter) Header() http.Header {
	return w.ResWriter.Header()
}

func (w *GzipWriter) WriteHeader(statusCode int) {
	if !strings.Contains(w.ResWriter.Header().Get("Content-Type"), "application/json") &&
		!strings.Contains(w.ResWriter.Header().Get("Content-Type"), "text/html") {
		w.ResWriter.WriteHeader(statusCode)
		w.GzipFlag = false
		return
	}
	w.GzipFlag = true
	w.ResWriter.Header().Set("Content-Encoding", "gzip")
	w.ResWriter.WriteHeader(statusCode)
}

func (w *GzipWriter) Write(b []byte) (int, error) {
	if !strings.Contains(w.ResWriter.Header().Get("Content-Type"), "application/json") &&
		!strings.Contains(w.ResWriter.Header().Get("Content-Type"), "text/html") {
		return w.ResWriter.Write(b)
	}
	return w.Writer.Write(b)
}

func (w *GzipWriter) Close() error {
	return w.Writer.Close()
}

func GzipCompressor(log *slog.Logger, compressorLevel int) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/gzip"),
		)
		log.Info("gzip compressor enabled")
		fn := func(w http.ResponseWriter, r *http.Request) {

			if !strings.Contains(strings.Join(r.Header.Values("Accept-Encoding"), " "), "gzip") {
				// If gzip is not supported then return uncompressed page.
				next.ServeHTTP(w, r)
				return
			}

			log.Info("gzip is supported")

			gzipWr, err := gzip.NewWriterLevel(w, compressorLevel)
			if err != nil {
				log.Error("failed to compress gzip")
				_, err := io.WriteString(w, err.Error())
				if err != nil {
					log.Error("failed to inform user")
					return
				}
				return
			}

			gz := &GzipWriter{ResWriter: w, Writer: gzipWr}
			next.ServeHTTP(gz, r)
			if gz.GzipFlag {
				err := gzipWr.Close()
				if err != nil {
					log.Error("failed to close gzip")
					_, err := io.WriteString(w, err.Error())
					if err != nil {
						log.Error("failed to inform user")
						return
					}
					return
				}
			}
		}
		return http.HandlerFunc(fn)
	}
}
