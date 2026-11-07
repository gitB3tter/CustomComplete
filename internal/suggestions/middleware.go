package suggestions

import (
	"bufio"
	"customcomplete/internal/textprocessing"
	"customcomplete/internal/transport"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func redirectCanonical(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		requestQuery := r.URL.Query()

		nextParam := requestQuery.Get("next")

		if nextParam == "" {
			next(w, r)
			return
		}

		requestQuery.Del("next")

		text := trimTrailingPrefix(r.URL.Query().Get("text")) + nextParam + " "

		requestQuery.Set("text", text)

		redirect := (&url.URL{
			Path:     "/suggestions",
			RawQuery: requestQuery.Encode(),
		})

		transport.SetCachePublic(w.Header(), time.Hour, 0)

		http.Redirect(w, r, redirect.String(), http.StatusMovedPermanently)
	}
}

func trimTrailingPrefix(text string) string {
	r, _ := utf8.DecodeLastRuneInString(text)
	if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
		// Ends in space, quote, punctuation, or is empty -> next-word mode, keep all of text!
		return text
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(textprocessing.ScanNormalizedWords)
	var lastWord string
	for scanner.Scan() {
		lastWord = scanner.Text()
	}
	return text[:len(text)-len(lastWord)]
}
