package mediatypes

import (
	"cmp"
	"customcomplete/internal/httpsyntax"
	"fmt"
	"slices"
	"strings"
)

// Media-Range w. Weight and order (as final (non-spec) pragmatic tiebreaker)
type acceptedRange struct {
	mRange
	// q-value, for simplicity as int >= 0 && <= 1000
	weight httpsyntax.Weight
	// position in the header
	order int
}

type AcceptedRanges []acceptedRange

// requires order of requested by specificity desc, number of params desc, order in header asc
func (requested AcceptedRanges) ExplicitlyAccepts(supported Type) bool {
	for _, acceptValue := range requested {
		if acceptValue.typ == "*" && acceptValue.mRange.subtype == "*" {
			continue
		}
		if supported.MatchedByRange(acceptValue) && acceptValue.weight > 0 {
			return true
		}
	}
	return false
}

func (ranges AcceptedRanges) String() string {
	var b strings.Builder
	for i, r := range ranges {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(r.typ)
		b.WriteByte('/')
		b.WriteString(r.subtype)
		if len(r.params) > 0 {
			b.WriteString("; ")
			b.WriteString(string(canonizeParams(r.params)))
		}
	}
	return b.String()

}

// parses content of the Accept Header into accepted Media Ranges
// see RFC 9110 §12.5.1
// orders by specificity desc, number of params desc, order in header asc
func ParseAccept(fields []string) AcceptedRanges {
	acceptValues := []acceptedRange{}
	if len(fields) == 0 {
		fields = []string{"*/*"}
	}

	order := 0
	for _, field := range fields {
		for _, segment := range httpsyntax.SplitList(field, ',') {
			if mediaRange, err := parseAcceptValue(segment, order); err == nil {
				order++
				acceptValues = append(acceptValues, mediaRange)
			}
		}
	}

	if len(acceptValues) > 1 {
		slices.SortStableFunc(acceptValues, func(a, b acceptedRange) int {

			if a.specificity() != b.specificity() {
				return cmp.Compare(b.specificity(), a.specificity())
			}

			if len(a.mRange.params) != len(b.mRange.params) {
				return cmp.Compare(len(b.mRange.params), len(a.mRange.params))
			}

			return cmp.Compare(a.order, b.order)
		})
	}

	return acceptValues
}

// Parses a Segment containing a media range and optionally parameters including weight/"q-value" into a typed object
func parseAcceptValue(segment string, position int) (acceptedRange, error) {

	var weight httpsyntax.Weight
	var ok bool

	segment, weight, ok = cutWeight(segment)
	if !ok {
		return acceptedRange{}, fmt.Errorf("invalid weight value")
	}

	mediaRange, err := parseMediaRange(segment)
	if err != nil {
		return acceptedRange{}, err
	}

	return acceptedRange{
		mRange: mediaRange,
		weight: weight,
		order:  position,
	}, nil
}

// splits the accept header by the weight ("q" parameter) and returns the weight-preceding mediarange with params + the weight as int 0-1000 (discarding accept-ext)
func cutWeight(field string) (mediaRange string, weight httpsyntax.Weight, ok bool) {
	for off, seg := range httpsyntax.SplitList(field, ';') {
		k, v, hasEq := strings.Cut(seg, "=")
		if !hasEq || !strings.EqualFold(strings.Trim(k, httpsyntax.HTTP_OWS), "q") {
			continue
		}
		//only ";q=..." no media range
		if off == 0 {
			return "", 0, false
		}
		weight, valid := httpsyntax.ParseWeight(strings.Trim(v, httpsyntax.HTTP_OWS))
		if !valid {
			return "", 0, false
		}
		//drop the weigth-preceding ';'
		return field[:off-1], weight, true
	}
	return field, 1000, true
}
