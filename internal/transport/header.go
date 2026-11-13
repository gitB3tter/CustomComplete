package transport

import (
	"customcomplete/internal/mediatypes"
	"net/http"
	"strings"
)

// appends a header name to Vary only if it is not already present.
func AddVaryOnce(h http.Header, value string) {
	addHeaderValueOnce(h, "Vary", value)
}

func SetAllowRead(w http.ResponseWriter) {
	w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodHead, http.MethodOptions}, ", "))
}

func SetContentType(w http.ResponseWriter, mt mediatypes.Type) {
	w.Header().Set("Content-Type", mt.String())
}

func addHeaderValueOnce(h http.Header, key, value string) {
	existing := h.Get(key)
	if existing == "" {
		h.Set(key, value)
		return
	}

	for item := range strings.SplitSeq(existing, ",") {
		if strings.EqualFold(strings.TrimSpace(item), value) {
			return
		}
	}

	h.Set(key, existing+", "+value)
}
