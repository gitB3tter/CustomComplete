package weblinking

import (
	"encoding/json"
	"strings"
	"testing"

	"golang.org/x/text/language"
)

func TestWebLinking(t *testing.T) {
	path := RootPath.Join("texts", "1")

	t.Run("ASCII title", func(t *testing.T) {
		l := New(Item, path).
			WithTitle("Works of Shakespeare")

		got := l.String()
		if !strings.Contains(got, `title="Works of Shakespeare"`) {
			t.Errorf("expected RFC 8187 title*=UTF-8'de'... in %q", got)
		}
	})

	t.Run("UTF-8 non-ASCII title + title language", func(t *testing.T) {
		l := New(Up, RootPath.Join("suggestions")).
			WithTitle("Vorschläge").
			WithTitleLang(language.German)

		got := l.String()
		if !strings.Contains(got, "title*=UTF-8'de'") {
			t.Errorf("expected RFC 8187 title*=UTF-8'de'... in %q", got)
		}
	})

	t.Run("RFC 9264 application/linkset+json structure", func(t *testing.T) {
		links := Links{
			New(Self, RootPath.Join("suggestions")),
			New(Next, RootPath.Join("suggestions")).
				WithQueryVal("text", "to be ").
				WithTitle("be"),
			New(Next, RootPath.Join("suggestions")).
				WithQueryVal("text", "to crack ").
				WithTitle("crack"),
		}

		raw, err := json.Marshal(links)
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed struct {
			Linkset []map[string]any `json:"linkset"`
		}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatalf("invalid JSON produced: %v", err)
		}
		if len(parsed.Linkset) != 1 {
			t.Fatalf("len(linkset) = %d, want 1", len(parsed.Linkset))
		}
		nextSlice, ok := parsed.Linkset[0]["next"].([]any)
		if !ok || len(nextSlice) != 2 {
			t.Errorf("linkset[0].next = %#v, want 2 items", parsed.Linkset[0]["next"])
		}
	})
}
