package textprocessing

import (
	"context"
	"customcomplete/internal/store"
	"fmt"
	"strings"
)

type inserQueriesConsts struct {
	start  []string
	params []string
}

var (
	queriesBatchMain  []string
	insertQueriesMain = inserQueriesConsts{
		start: []string{
			1:  "INSERT OR REPLACE INTO ngram1(source_id,word1, occurrences) VALUES ",
			2:  "INSERT OR REPLACE INTO ngram2(source_id,word1, word2, occurrences) VALUES ",
			3:  "INSERT OR REPLACE INTO ngram3(source_id,word1, word2, word3, occurrences) VALUES ",
			4:  "INSERT OR REPLACE INTO ngram4(source_id,word1, word2, word3, word4, occurrences) VALUES ",
			5:  "INSERT OR REPLACE INTO ngram5(source_id,word1, word2, word3, word4, word5, occurrences) VALUES ",
			6:  "INSERT OR REPLACE INTO ngram6(source_id,word1, word2, word3, word4, word5, word6, occurrences) VALUES ",
			7:  "INSERT OR REPLACE INTO ngram7(source_id,word1, word2, word3, word4, word5, word6, word7, occurrences) VALUES ",
			8:  "INSERT OR REPLACE INTO ngram8(source_id,word1, word2, word3, word4, word5, word6, word7, word8, occurrences) VALUES ",
			9:  "INSERT OR REPLACE INTO ngram9(source_id,word1, word2, word3, word4, word5, word6, word7, word8, word9, occurrences) VALUES ",
			10: "INSERT OR REPLACE INTO ngram10(source_id,word1,word2, word3, word4, word5, word6, word7, word8, word9, word10, occurrences) VALUES ",
		},
		params: []string{
			1:  "(?,?,?)",
			2:  "(?,?,?,?)",
			3:  "(?,?,?,?,?)",
			4:  "(?,?,?,?,?,?)",
			5:  "(?,?,?,?,?,?,?)",
			6:  "(?,?,?,?,?,?,?,?)",
			7:  "(?,?,?,?,?,?,?,?,?)",
			8:  "(?,?,?,?,?,?,?,?,?,?)",
			9:  "(?,?,?,?,?,?,?,?,?,?,?)",
			10: "(?,?,?,?,?,?,?,?,?,?,?,?)",
		},
	}
)

func init() {
	var b strings.Builder
	b.Grow(len(insertQueriesMain.start[10]) + (len(insertQueriesMain.params[10])+1)*paramLimit/11)
	paramsPerRow := 2 //all contain occurrences, sourceid
	queriesBatchMain = []string{0: ""}
	for ngramLength := 1; ngramLength <= 10; ngramLength++ {
		b.Reset()
		paramsPerRow++
		b.WriteString(insertQueriesMain.start[ngramLength])
		numbRows := paramLimit / paramsPerRow
		b.WriteString(insertQueriesMain.params[ngramLength])
		for row := 1; row < numbRows; row++ {
			b.WriteByte(',')
			b.WriteString(insertQueriesMain.params[ngramLength])
		}
		queriesBatchMain = append(queriesBatchMain, b.String())
	}
}

func ResetInterruptedJobs(db store.ReadWrite) error {
	_, err := db.Exec(`
	UPDATE source
	SET status = 'pending'
	WHERE status = 'processing';
	`)
	return err
}

func getContent(ctx context.Context, sourceId int64, db store.Read) (content []byte, err error) {
	if err = db.QueryRowContext(ctx, `SELECT content
	FROM source
	WHERE id=?`, sourceId).Scan(&content); err != nil {
		return nil, fmt.Errorf("failed to retrieve filecontent: %w: %w", ErrFatal, err)
	}
	return content, nil
}
