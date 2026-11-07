package textprocessing

import (
	"bufio"
	"bytes"
	"context"
	"customcomplete/internal/store"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

var (
	queryTokenStart = "INSERT INTO token (word, folded) VALUES "
	queryTokenEnd   = " ON CONFLICT(word) DO UPDATE SET word=excluded.word RETURNING id, word"
	placeholder2    = "(?,?)"
	queryTokenBatch string
)

func init() {
	var b strings.Builder
	b.Grow(len(queryTokenStart) + (paramLimit/2)*(len(placeholder2)+1) + len(queryTokenEnd))
	b.WriteString(queryTokenStart)
	b.WriteString(placeholder2)
	for i := 0; i < paramLimit-2; i += 2 {
		b.WriteByte(',')
		b.WriteString(placeholder2)
	}
	b.WriteString(queryTokenEnd)
	queryTokenBatch = b.String()

}

func estimatedWords(lenBytes int) int {
	return int(lenBytes / 6) // total words (~916K for 5.5MB shakespeare)
}

func estmatedUnique(wordCount int) int {
	return int(100 * math.Sqrt(float64(wordCount)))
}

// wordcount = unique word count (not/half of wordInsert (word + folded & normalized word))
func prepareStatementsTokens(ctx context.Context, tx *sql.Tx, wordcount int) (stmtBatch *sql.Stmt, stmtRemaining *sql.Stmt, err error) {
	rowsPerBatch := paramLimit / 2
	remainingRows := wordcount % rowsPerBatch
	if remainingRows > 0 {
		var b strings.Builder
		b.Grow(len(queryTokenStart) + (remainingRows)*(len(placeholder2)+1) + len(queryTokenEnd))
		b.WriteString(queryTokenStart)
		b.WriteString(placeholder2)
		for i := 0; i < remainingRows-1; i++ {
			b.WriteByte(',')
			b.WriteString(placeholder2)
		}
		b.WriteString(queryTokenEnd)
		if stmtRemaining, err = tx.PrepareContext(ctx, b.String()); err != nil {
			return nil, nil, err
		}
	}
	if wordcount >= rowsPerBatch {
		if stmtBatch, err = tx.PrepareContext(ctx, queryTokenBatch); err != nil {
			return nil, nil, err
		}
	}

	return stmtBatch, stmtRemaining, nil
}

func NFKCFold(s string) string {
	return norm.NFKC.String(cases.Fold().String(norm.NFKC.String(s)))
}

func createSliceInsertTokens(words map[string]int64) []any {
	wordsInsert := make([]any, 0, len(words)*2)
	for word := range words {
		wordsInsert = append(wordsInsert, word, NFKCFold(word))
	}
	return wordsInsert
}

func insertTokens(ctx context.Context, db store.ReadWrite, tokens map[string]int64) (map[string]int64, error) {

	wordsInsert := createSliceInsertTokens(tokens)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmtBatch, stmtRemaining, err := prepareStatementsTokens(ctx, tx, len(tokens))
	if err != nil {
		return nil, err
	}
	if stmtBatch != nil {
		defer stmtBatch.Close()
	}
	if stmtRemaining != nil {
		defer stmtRemaining.Close()
	}

	start := 0
	var id int64
	var word string
	for end := paramLimit; end <= len(wordsInsert); end += paramLimit {
		rows, err := stmtBatch.QueryContext(ctx, wordsInsert[start:end]...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute statement insert tokens batch: %w", err)
		}
		for rows.Next() {
			if err := rows.Scan(&id, &word); err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to insert tokens batch: %w: %w", ErrFatal, err)
			}
			tokens[word] = id
		}
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("error while inserting tokens batch: %w: %w", ErrFatal, err)
		}
		start = end
		rows.Close()
	}

	if start < len(wordsInsert) {
		rows, err := stmtRemaining.QueryContext(ctx, wordsInsert[start:]...)
		if err != nil {

			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			if err := rows.Scan(&id, &word); err != nil {
				return nil, fmt.Errorf("failed to insert tokens remaining: %w: %w", ErrFatal, err)
			}
			tokens[word] = id
		}
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("error while inserting tokens remaining: %w: %w", ErrFatal, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return tokens, nil
}

func generateTokens(ctx context.Context, sourceID int64, readWrite store.ReadWrite, read store.Read, progress reportProgress) ([]int64, error) {

	content, err := getContent(ctx, sourceID, read)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))

	scanner.Split(ScanNormalizedWords)

	estWords := estimatedWords(len(content))
	words := make([]string, 0, estWords)
	tokenMapping := make(map[string]int64, estmatedUnique(estWords))

	var word string

	for scanner.Scan() {
		b := scanner.Bytes()
		if _, exists := tokenMapping[string(b)]; !exists {
			word = string(b)
			tokenMapping[word] = 0 //filled in after insert
		} else {
			word = string(b)
		}
		words = append(words, word)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	progress(50)

	tokenMapping, err = insertTokens(ctx, readWrite, tokenMapping)
	if err != nil {
		return nil, err
	}

	progress(80)

	tokens := make([]int64, 0, len(words))

	for _, word := range words {
		tokens = append(tokens, tokenMapping[word])
	}

	return tokens, nil
}
