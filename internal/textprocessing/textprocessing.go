package textprocessing

import (
	"context"
	"customcomplete/internal/i18n"
	"customcomplete/internal/store"
	_ "embed"
	"errors"
)

func Init(db store.ReadWrite) error {
	return ResetInterruptedJobs(db)
}

// report progress of the function from 1-100
type reportProgress func(int)

type Broker interface {
	Update(id int64, percentage int, message i18n.Key, args ...any)
}

// round number below real sqliteParamLimit of 32766 (in modern sqlite)
const paramLimit = 10000

// sqliteQueryLimit = 1000000000
// const QueryLimit = 1000000

// to reduce needless syntax bc of ngram-maps keyed by different length
type BatchSpec struct {
	Handler func(func([]int64, int64))
}

var (
	//errors
	ErrFatal = errors.New("Fatal processing error: Worker should stop")
	//procsessing Status
)

func process(ctx context.Context, sourceId int64, readWrite store.ReadWrite, read store.Read, broker Broker) (int, error) {
	broker.Update(sourceId, 0, i18n.ProcessingReadingText)
	tokens, err := generateTokens(ctx, sourceId, readWrite, read, func(n int) {
		broker.Update(sourceId, n/50, i18n.ProcessingReadingText)
	})
	if err != nil {
		return 0, err
	}
	broker.Update(sourceId, 2, i18n.ProcessingCounting)

	wordCount := len(tokens)

	counts := getCounts(tokens, func(n int) {
		broker.Update(sourceId, 2+n*8/100, i18n.ProcessingCounting)
	})

	broker.Update(sourceId, 10, i18n.ProcessingPreparingSuggestions)

	sorted := getSorted(counts, func(n int) {
		broker.Update(sourceId, 10+(n*3/2)/10, i18n.ProcessingPreparingSuggestions)
	})

	broker.Update(sourceId, 25, i18n.ProcessingSavingSuggestions)

	//scratchDB, countSugg, err := insertScratch(ctx, getSorted(getCounts(tokens)), sourceId)
	err = insert(ctx, readWrite, sorted, sourceId, func(n int) {
		broker.Update(sourceId, 25+(n*74)/100, i18n.ProcessingSavingSuggestions)
	})
	if err != nil {
		return wordCount, err
	}

	return wordCount, nil
}
