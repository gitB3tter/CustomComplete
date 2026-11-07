package transport

import (
	"context"
	"customcomplete/internal/httpsyntax"
	"net/http"
	"strings"
)

type Preference string

type Preferences []Preference

type contextKey string

const (
	PreferenceKey        contextKey = "prefer"
	ReturnMinimal        Preference = "return=minimal"
	ReturnRepresentation Preference = "return=representation"
	HandlingLenient      Preference = "handling=lenient"
	HandlingStrict       Preference = "handling=strict"
)

func ParsePreferHeader(headerVal string) Preferences {
	segments := httpsyntax.SplitList(headerVal, ',')
	var preferences Preferences
	for _, segment := range segments {
		if prefer, ok := parsePreferSegment(segment); ok {
			preferences = append(preferences, prefer)
		}
	}
	return preferences
}

func parsePreferSegment(segment string) (Preference, bool) {
	//currently no preference w. possibly escaped ';' nor params supported
	preference, _, _ := strings.Cut(segment, ";")
	key, val, hasVal := strings.Cut(preference, "=")
	key = strings.ToLower(strings.TrimSpace(key))
	val = strings.Trim(strings.TrimSpace(val), "\"")
	if !hasVal {
		return "", false
	}
	switch key {
	case "return":
		if val == "minimal" {
			return ReturnMinimal, true
		}
		if val == "representation" {
			return ReturnRepresentation, true
		}
	case "handling":
		if val == "lenient" {
			return HandlingLenient, true
		}
		if val == "strict" {
			return HandlingStrict, true
		}

	}

	return "", false
}

func WithPreferences(r *http.Request, preferences Preferences) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), PreferenceKey, preferences))
}

func GetPreferences(ctx context.Context) (Preferences, bool) {
	preferences, ok := ctx.Value(PreferenceKey).(Preferences)
	return preferences, ok
}

func ApplyPreferences(h http.Header, preferences ...Preference) {
	for _, preference := range preferences {
		h.Add("Preference-Applied", string(preference))
	}
}

func ClearPreferences(h http.Header) {
	h.Del("Preference-Applied")
}
