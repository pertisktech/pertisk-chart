package api

import (
	"net/http"

	"github.com/CAFxX/httpcompression"
)

// GetCompressedHandler wraps the HTTP handler with compression middleware
// Supports zstd, brotli, gzip, and deflate based on client support
// DefaultAdapter includes zstd compression by default
func GetCompressedHandler(handler http.Handler) (http.Handler, error) {
	adapter, err := httpcompression.DefaultAdapter()
	if err != nil {
		return nil, err
	}
	return adapter(handler), nil
}

