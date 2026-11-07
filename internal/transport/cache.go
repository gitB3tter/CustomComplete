package transport

import (
	"fmt"
	"net/http"
	"time"
)

func SetCachePublic(h http.Header, maxAge time.Duration, staleWhileRevalidate time.Duration) {
	ageSeconds := int(maxAge.Seconds())
	if staleWhileRevalidate > 0 {
		h.Set("Cache-Control", fmt.Sprintf("public, max-age=%d, stale-while-revalidate=%d", ageSeconds, int(staleWhileRevalidate.Seconds())))
		return
	}
	h.Set("Cache-Control", fmt.Sprintf("public, max-age=%d", ageSeconds))
}

func SetCacheImmutable(h http.Header) {
	h.Set("Cache-Control", "public, max-age=31536000, immutable")
}

func SetNoCache(h http.Header) {
	h.Set("Cache-Control", "no-cache")
}

// advises the client not to cache (dont confuse with directive 'no-cache')
func SetDontCache(h http.Header) {
	h.Set("Cache-Control", "no-store, max-age=0")
}
