package transport

import (
	"bytes"
	"encoding/json"
)

type JSONType string

const (
	JSONTypeObject  JSONType = "object"
	JSONTypeArray   JSONType = "array"
	JSONTypeString  JSONType = "string"
	JSONTypeNumber  JSONType = "number"
	JSONTypeBoolean JSONType = "boolean"
	JSONTypeNull    JSONType = "null"
	JSONTypeInvalid JSONType = "invalid"
)

func JsonType(raw json.RawMessage) JSONType {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return JSONTypeInvalid
	}

	switch trimmed[0] {
	case '{':
		return JSONTypeObject
	case '[':
		return JSONTypeArray
	case '"':
		return JSONTypeString
	case 't', 'f':
		return JSONTypeBoolean
	case 'n':
		return JSONTypeNull
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return JSONTypeNumber
	default:
		return JSONTypeInvalid
	}
}
