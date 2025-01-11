package middlewares

import (
	"net/http"
	"strings"

	compress "github.com/Melikhov-p/go-loyalty-system/internal/compressor"
	"go.uber.org/zap"
)

func (m *Middleware) GzipMiddleware(h http.Handler) http.Handler {
	comp := func(w http.ResponseWriter, r *http.Request) {
		ow := w

		content := w.Header().Get("Content-Type")

		if content == "application/json" || content == "text/html" {
			acceptEncoding := r.Header.Get("Accept-Encoding")
			if strings.Contains(acceptEncoding, "gzip") {
				cw := compress.NewCompressWrite(w)
				ow = cw
				defer func() {
					if err := cw.Close(); err != nil {
						m.logger.Error("error closing compressWriter", zap.Error(err))
					}
				}()
			}
		}

		contentEncoding := r.Header.Get("Content-Encoding")

		if strings.Contains(contentEncoding, "gzip") {
			cr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				m.logger.Error("error getting compress reader", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func() {
				if err = cr.Close(); err != nil {
					m.logger.Error("error closing compressReader", zap.Error(err))
				}
			}()
		}

		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(comp)
}
