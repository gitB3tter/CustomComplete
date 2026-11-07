package catalog

import (
	"bytes"
	"crypto/sha256"
	"customcomplete/internal/constraints"
	"customcomplete/internal/errs"
	"encoding/hex"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type sourceID int64

type SourceStatus string

const (
	StatusPending    SourceStatus = "pending"
	StatusProcessing SourceStatus = "processing"
	StatusComplete   SourceStatus = "completed"
	StatusFailed     SourceStatus = "failed"
)

type SourceContents []byte

func ParseSourceContents(b []byte) (SourceContents, *errs.CodedError) {

	if len(b) > int(MaxFileSize) {
		return nil, errs.MaxBytes(MaxFileSize)
	}

	if len(bytes.TrimSpace(b)) == 0 {
		return nil, errs.Blank()
	}

	if !utf8.Valid(b) {
		return nil, errs.InvalidUTF8()
	}

	if !constraints.WordCountMin(string(b), MinWordsContents) {
		return nil, errs.MinWords(MinWordsContents, constraints.WordCount(string(b)))
	}

	return SourceContents(b), nil
}

// sha256 or smth
type SourceContentHash string

func getSourceContenthash(contents []byte) SourceContentHash {
	//dont destroy/corrupt the passed in contents
	trimmed := bytes.TrimSpace(contents)
	workingCopy := make([]byte, len(trimmed))
	copy(workingCopy, trimmed)
	hashArray := sha256.Sum256(collapseSpace(norm.NFC.Bytes(workingCopy)))
	return SourceContentHash(hex.EncodeToString(hashArray[:]))
}

func (service *Service) GetSourceContents(docID DocumentID) (content SourceContents, err error) {
	return service.store.getContents(docID)
}

// Collapses consecutive whitespace (spaces, tabs, newlines) into a single space in-place.
func collapseSpace(b []byte) []byte {
	n := 0
	lastWasSpace := false
	for _, v := range b {
		isSpace := v == ' ' || v == '\n' || v == '\t' || v == '\r'
		if isSpace {
			if !lastWasSpace {
				b[n] = ' '
				n++
			}
			lastWasSpace = true
		} else {
			b[n] = v
			n++
			lastWasSpace = false
		}
	}
	return b[:n]
}
