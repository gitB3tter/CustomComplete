package textprocessing

import (
	"context"
	"customcomplete/internal/store"
	"database/sql"
	"fmt"
	"strings"
)

func insert(ctx context.Context, db store.ReadWrite, sorted ngrams, sourceId int64, progress reportProgress) (err error) {

	weights := [...]float64{
		0, // no table 0
		0.25, 0.5, 1.5, 2, 2.25,
		2.25, 2.5, 2.5, 2.75, 3,
	}

	progressStep := func(done int, factor float64) reportProgress {
		return func(n int) {
			progress(done + int(factor*float64(n)/100)*5)
		}
	}

	fail := func(table string, err error) error {
		return fmt.Errorf("failed to insert %s into main: %w", table, err)
	}
	done := 0.0
	step := 1
	if err = insertTable(ctx, db, 1, sorted.words, sourceId, func(a []any, r [2]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("words", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++
	if err = insertTable(ctx, db, 2, sorted.ngram2, sourceId, func(a []any, r [3]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("2grams", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++
	if err = insertTable(ctx, db, 3, sorted.ngram3, sourceId, func(a []any, r [4]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("3grams", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++
	if err = insertTable(ctx, db, 4, sorted.ngram4, sourceId, func(a []any, r [5]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("4grams", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++
	if err = insertTable(ctx, db, 5, sorted.ngram5, sourceId, func(a []any, r [6]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("5grams", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++
	if err = insertTable(ctx, db, 6, sorted.ngram6, sourceId, func(a []any, r [7]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("6grams", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++
	if err = insertTable(ctx, db, 7, sorted.ngram7, sourceId, func(a []any, r [8]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("7grams", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++
	if err = insertTable(ctx, db, 8, sorted.ngram8, sourceId, func(a []any, r [9]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("8grams", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++
	if err = insertTable(ctx, db, 9, sorted.ngram9, sourceId, func(a []any, r [10]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("9grams", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++
	if err = insertTable(ctx, db, 10, sorted.ngram10, sourceId, func(a []any, r [11]int64, sourceId int64) []any { return appendIntsMain(a, r[:], sourceId) }, progressStep(int(done), weights[step])); err != nil {
		return fail("10grams", err)
	}
	done = done + weights[step]*5
	progress(int(done))
	step++

	return err

}

const batchesPerTx = 50

func insertTable[Row any](
	ctx context.Context,
	db store.ReadWrite,
	ngramWidth int,
	rows []Row,
	sourceID int64,
	appendRow func(args []any, row Row, sourceID int64) []any,
	progress reportProgress,
) error {
	if len(rows) == 0 {
		progress(100)
		return nil
	}

	rowWidth := ngramWidth + 2
	rowsPerBatch := paramLimit / rowWidth
	batchsize := rowsPerBatch * rowWidth
	batchesTotal := (len(rows) + rowsPerBatch - 1) / rowsPerBatch

	var (
		tx        *sql.Tx
		stmtBatch *sql.Stmt
		err       error
	)

	batchCount, batchesDone := 0, 0

	closeTx := func(commit bool, batchesDone int) error {
		if stmtBatch != nil {
			stmtBatch.Close()
			stmtBatch = nil
		}
		if tx == nil {
			return nil
		}
		var txErr error
		if commit {
			txErr = tx.Commit()
			if txErr == nil {
				progress((batchesDone * 100) / batchesTotal)
			}
		} else {
			txErr = tx.Rollback()
		}
		tx = nil
		batchCount = 0
		return txErr
	}
	defer func() { _ = closeTx(false, 0) }()

	args := make([]any, 0, batchsize)
	for _, row := range rows {
		args = appendRow(args, row, sourceID)
		if len(args) == batchsize {
			if tx == nil {
				if tx, err = db.BeginTx(ctx, nil); err != nil {
					return err
				}
				if stmtBatch, err = tx.PrepareContext(ctx, queriesBatchMain[ngramWidth]); err != nil {
					return err
				}
			}
			if _, err = stmtBatch.ExecContext(ctx, args...); err != nil {
				return err
			}
			args = args[:0]
			batchCount++
			batchesDone++
			if batchCount >= batchesPerTx {
				if err = closeTx(true, batchesDone); err != nil {
					return err
				}
			}
		}
	}

	if len(args) > 0 {
		if tx == nil {
			if tx, err = db.BeginTx(ctx, nil); err != nil {
				return err
			}
		}
		var b strings.Builder
		b.Grow(len(insertQueriesMain.start[ngramWidth]) + len(args)*2)
		b.WriteString(insertQueriesMain.start[ngramWidth])

		numbRows := len(args) / rowWidth
		b.WriteString(insertQueriesMain.params[ngramWidth])
		for row := 1; row < numbRows; row++ {
			b.WriteByte(',')
			b.WriteString(insertQueriesMain.params[ngramWidth])
		}
		if _, err := tx.ExecContext(ctx, b.String(), args...); err != nil {
			return err
		}
		batchesDone++
	}
	return closeTx(true, batchesDone)
}

func appendIntsMain(a []any, nums []int64, sourceID int64) []any {
	a = append(a, sourceID)
	for _, v := range nums {
		a = append(a, v)
	}
	return a
}
