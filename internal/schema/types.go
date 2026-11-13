package schema

import (
	"bytes"
	"encoding/json"
)

type Type string

const (
	Object Type = "object"
	Array  Type = "array"
	String Type = "string"
	Number Type = "number"
	//json schema only
	Integer Type = "integer"
	Boolean Type = "boolean"
	Null    Type = "null"
	Invalid Type = "invalid"
)

func JsonType(raw json.RawMessage) Type {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return Invalid
	}

	switch trimmed[0] {
	case '{':
		return Object
	case '[':
		return Array
	case '"':
		return String
	case 't', 'f':
		return Boolean
	case 'n':
		return Null
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if bytes.ContainsAny(trimmed, ".eE") {
			return Number
		}
		return Integer
	default:
		return Invalid
	}
}
