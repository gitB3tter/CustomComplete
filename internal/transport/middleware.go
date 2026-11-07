package transport

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/logging"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/weblinking"
	"net/http"
	"net/url"
)

func MB(mb int) (byte int64) {
	return int64(mb * 1_000_000)
}

// maxBodySize should include a small buffer for handler rejecting only slightly too large uploads with error in context
func LimitBody(next http.HandlerFunc, accepted mediatypes.Types, maxBodySize int64) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		langTag := i18n.LangTagFromContext(r.Context())
		T := i18n.Func(langTag)

		hasBody := r.ContentLength != 0 || len(r.TransferEncoding) > 0
		if len(accepted) == 0 || maxBodySize == 0 {
			if hasBody {
				BodyNotAllowed(w, r, langTag)
				return
			}
			r.Body = http.NoBody
			next.ServeHTTP(w, r)
			return
		}

		linkHome := weblinking.New(weblinking.Up, &url.URL{Path: "/"}).WithTitle(T(i18n.NavigationHome))

		if r.ContentLength > maxBodySize {
			ContentTooLarge(w, r, i18n.FormatBytesLocale(maxBodySize, langTag), linkHome)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

		contentType, err := mediatypes.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			UnsupportedMediatype(w, r, linkHome, contentType, accepted)
			return
		}

		if hasBody && !accepted.Supports(contentType) {
			UnsupportedMediatype(w, r, linkHome, contentType, accepted)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func NoBody(next http.HandlerFunc) http.HandlerFunc {
	return LimitBody(next, mediatypes.Types{}, 0)
}

func NegotiateLanguague(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tag := i18n.ResolveFromHeader(r.Header.Get("Accept-Language"))

		AddVaryOnce(w.Header(), "Accept-Language")

		w.Header().Set("Content-Language", tag.String())
		next.ServeHTTP(w, r.WithContext(i18n.WithLangTag(r.Context(), tag)))
	})
}

type logWriter struct {
	http.ResponseWriter
	status int
}

func (w *logWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *logWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = logging.StartLog(r)
		lw := &logWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)
		logging.EndLog(r, lw.status)
	})
}

func ParsePreference(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, WithPreferences(r, ParsePreferHeader(r.Header.Get("Prefer"))))
	})
}
