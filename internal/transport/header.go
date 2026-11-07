package transport

import (
	"customcomplete/internal/mediatypes"
	"net/http"
	"strings"
)

// appends a header name to Vary only if it is not already present.
func AddVaryOnce(h http.Header, name string) {
	existing := h.Get("Vary")
	if existing == "" {
		h.Set("Vary", name)
		return
	}

	for item := range strings.SplitSeq(existing, ",") {
		if strings.EqualFold(strings.TrimSpace(item), name) {
			return
		}
	}

	h.Set("Vary", existing+", "+name)
}

func SetAllowRead(w http.ResponseWriter) {
	w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodHead, http.MethodOptions}, ", "))
}

func SetContentType(w http.ResponseWriter, mt mediatypes.Type) {
	w.Header().Set("Content-Type", mt.String())
}
