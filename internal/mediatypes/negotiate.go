package mediatypes

import (
	"context"
	"customcomplete/internal/httpsyntax"
)

type contextKey string

const (
	NegotiatedKey contextKey = "accept"
)

func SetNegotiated(ctx context.Context, ResponseType Type) context.Context {
	return context.WithValue(ctx, NegotiatedKey, ResponseType)
}

func GetNegotiated(ctx context.Context) (Type, bool) {
	negotiated, ok := ctx.Value(NegotiatedKey).(Type)
	return negotiated, ok
}

// Negotiates the media type of the supported ones to serve based on the requested Media Ranges
// requires AcceptedRanges ordered by specificity, number of params and order desc
func (supported Types) Negotiate(requested AcceptedRanges) (Type, bool) {

	var matchedType Type
	matchedWeight := httpsyntax.WeightZero
	matchedClientOrder := len(requested) + 1

supported:
	for _, supportedType := range supported {
		for _, acceptValue := range requested {
			if supportedType.MatchedByRange(acceptValue) {
				if acceptValue.weight < matchedWeight {
					continue supported
				}
				// (server preference is to apply) client header order as final tiebreaker
				if acceptValue.weight == matchedWeight && acceptValue.order >= matchedClientOrder {
					continue supported
				}
				matchedType = supportedType
				matchedWeight = acceptValue.weight
				matchedClientOrder = acceptValue.order
				continue supported // all subsequent accept values match w. lower specificity
			}
		}
	}
	if matchedWeight > httpsyntax.WeightZero {
		return matchedType, true
	}
	return Type{}, false
}

// returns if the provided mediaType can be processed given the supported/processable MediaTypes for this route
func (supported Types) Supports(provided Type) bool {
	if provided.typ == typeMultipart {
		if _, ok := provided.params.Get("boundary"); !ok {
			return false
		}
	}
checkType:
	for _, supportedType := range supported {
		if provided.typ != supportedType.typ ||
			provided.subtype != supportedType.subtype {
			continue
		}

		if supportedType.params == provided.params {
			return true
		}

		//server doesnt restrict params, client doesnt need to provide boundary
		if supportedType.params == "" && provided.typ != typeMultipart {
			return true
		}

		for nameProvided, valueProvided := range provided.params.All() {
			if valueSupported, ok := supportedType.params.Get(nameProvided); ok && valueProvided != valueSupported {
				continue checkType
			}
		}
		return true
	}
	return false
}

func (supported Type) MatchedByRange(requested acceptedRange) bool {

	if requested.typ == "*" && requested.weight > 0 {
		return true
	}

	if requested.typ != "*" && requested.typ != supported.typ ||
		requested.subtype != "*" && requested.subtype != supported.subtype {
		return false
	}

	if len(requested.params) == 0 {
		return true
	}

	if supported.params == "" {
		return false
	}

	matchedCount := 0
	for name, value := range supported.params.All() {
		if requestedValue, ok := requested.params[name]; ok {
			if requestedValue != value {
				return false
			}
			matchedCount++
		}
	}

	return matchedCount == len(requested.params)
}
