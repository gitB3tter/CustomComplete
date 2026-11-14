package catalog

import (
	"bytes"
	"customcomplete/internal/constraints"
	"customcomplete/internal/errs"
	"customcomplete/internal/schema"
	"unicode/utf8"
)

type sourceID int64

type SourceStatus string

const (
	StatusPending    SourceStatus = "pending"
	StatusProcessing SourceStatus = "processing"
	StatusComplete   SourceStatus = "completed"
	StatusFailed     SourceStatus = "failed"
)

type SourceContents constraints.FileContents

func ParseSourceContents(contents []byte) (SourceContents, error) {

	fileContents, err := constraints.ParseFileContents(contents)
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(fileContents)) == 0 {
		return nil, new(errs.Blank())
	}

	if !utf8.Valid(contents) {
		return nil, new(errs.InvalidUTF8())
	}

	if !constraints.WordCountMin(string(contents), MinWordsContents) {
		return nil, new(errs.MinWords(MinWordsContents, constraints.WordCount(string(contents))))
	}

	return SourceContents(contents), nil
}

func (SourceContents) Schema() schema.Schema {
	s := new(constraints.FileContents).Schema()
	s.XConstraints.MinBytes = int64(MinWordsContents + 1)
	s.XConstraints.MinWords = MinWordsContents
	s.XConstraints.Encoding = "utf-8"
	return s
}

func (service *Service) GetSourceContents(docID DocumentID) (content SourceContents, err error) {
	return service.store.getContents(docID)
}
