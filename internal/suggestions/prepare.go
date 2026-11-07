package suggestions

import (
	"customcomplete/internal/store"
	"database/sql"
	"fmt"
	"math"
	"strings"
)

const (

	// Standard n-gram models usually cap at 3-5 words because longer contexts are increasingly all unique, rarely match and mostly overfit. In contrast, this goes, by default, up to 9 context words (last word of 10gram as "next") just because/for fun, relying on preferredNgramLength + lowCountDampeningBase to keep long matches from turning suggestions into verbatim text recitation
	maxContextWords = 9

	// Heuristic tuning knobs for scoring ngram candidates.
	// good-feel tuned against classic literature (Shakespeare, Moby Dick, Sherlock Holmes)
	// when keep taking the first suggestion, the higher both values the more the original source gets parroted, the lower both values the more it flips between sources or passages within one

	// peak of the bell-curve weight for n-gram size (n).
	// Lower values favor generic short matches, higher values favor verbatim logn quotes.
	// preferably keep between 3.0 and 4.0 as to favor 3- and 4-grams
	preferredNgramLength = 3.6

	// distrust of rare, long n-grams.
	// added to occurrences denominator as pow(base, n-2) so that especially long ngrams occuring only do not overpower common shorter ngrams
	// the higher the value, the more occurrences are required before a long ngram gets a high score
	lowCountDampingBase = 1.3

	// Zero-context queries (unigram frequency fallback when no prior words exist).

	queryAutocompleteAllZero = `
WITH src AS MATERIALIZED (
SELECT DISTINCT d.source_id
FROM document d
JOIN source s ON s.id = d.source_id
WHERE s.status = 'completed'
)
SELECT
t.word
FROM token t
JOIN ngram1 n ON n.word1 = t.id
WHERE t.folded >= ?1 AND t.folded < ?2
AND n.source_id IN src
GROUP BY t.id, t.word
ORDER BY (1.0 * SUM(n.occurrences) * COUNT(*)) DESC
LIMIT ?;
`

	queryAutocompleteDocZero = `
WITH src AS MATERIALIZED (
SELECT d.source_id
FROM document d
JOIN source s ON s.id = d.source_id
WHERE d.id = ?1 AND s.status = 'completed'
)
SELECT
t.word
FROM token t
JOIN ngram1 n ON n.word1 = t.id
WHERE t.folded >= ?2 AND t.folded < ?3
  AND n.source_id IN src
GROUP BY t.id, t.word
ORDER BY SUM(n.occurrences) DESC
LIMIT ?;
`

	queryNextWordAllZero = `
WITH src AS MATERIALIZED (
SELECT DISTINCT d.source_id
FROM document d
JOIN source s ON s.id = d.source_id
WHERE s.status = 'completed'
)
SELECT
t.word
FROM (
SELECT
word1 AS token_id,
(1.0 * SUM(occurrences) * COUNT(*)) AS score
FROM ngram1
WHERE source_id IN src
GROUP BY word1
ORDER BY score DESC
LIMIT ?
) candidates
JOIN token t ON t.id = candidates.token_id
ORDER BY candidates.score DESC;
`

	queryNextWordDocZero = `
WITH src AS MATERIALIZED (
	SELECT d.source_id
	FROM document d
	JOIN source s ON s.id = d.source_id
	WHERE d.id = ?1 AND s.status = 'completed'
)
SELECT
	t.word
FROM ngram1 n
JOIN token t ON t.id = n.word1
WHERE n.source_id IN src
GROUP BY t.id, t.word
ORDER BY SUM(n.occurrences) DESC
LIMIT ?;
`

	// SQL fragments for dynamically building 1..maxContextWords queries.

	cteContext        = `w%[1]d AS MATERIALIZED (SELECT id FROM token WHERE folded = ?%[1]d)`
	cteDocumentFilter = `
src AS MATERIALIZED (
SELECT d.source_id
FROM document d
JOIN source s ON s.id = d.source_id
WHERE d.id = ?%[1]d AND s.status = 'completed'
)`

	cteStatus = `
src AS MATERIALIZED (
SELECT DISTINCT d.source_id
FROM document d
JOIN source s ON s.id = d.source_id
WHERE s.status = 'completed'
)`

	queryStart = `
SELECT
next.word
FROM (
`

	// ngram scores candidates across all documents (with cross-source diversity multiplier).
	// Args: [1] n, [2] weight, [3] n-1, [4] whereClause, [5] damping
	ngram = `
SELECT
word%[1]d AS token_id,
%[2]g
* (1.0 * SUM(occurrences) / (SUM(SUM(occurrences)) OVER () + %[5]g))
* (1.0 * COUNT(DISTINCT source_id) / (
(SELECT COUNT(DISTINCT source_id) FROM ngram%[3]d
WHERE %[4]s) + 1.0)
) AS score
FROM ngram%[1]d
WHERE %[4]s
GROUP BY word%[1]d
`

	// ngramDoc scores candidates within a single document (without source-diversity multiplier).
	// Args: [1] n, [2] weight, [3] whereClause, [4] damping
	ngramDoc = `
SELECT
word%[1]d AS token_id,
%[2]g
* (1.0 * SUM(occurrences) / (SUM(SUM(occurrences)) OVER () + %[4]g)) AS score
FROM ngram%[1]d
WHERE %[3]s
GROUP BY word%[1]d
`

	unionAll = `
	UNION ALL
	`

	where = `word%[1]d IN w%[2]d`

	whereSource = `source_id IN src`

	and = ` AND `

	autocompleteQueryEnd = `
) candidates
JOIN token next ON next.id = candidates.token_id
WHERE next.folded >= ?%[1]d AND next.folded < ?%[2]d
GROUP BY next.id, next.word
ORDER BY SUM(candidates.score) DESC
LIMIT ?%[3]d;
`

	nextWordQueryEnd = `
) candidates
JOIN token next ON next.id = candidates.token_id
GROUP BY next.id, next.word
ORDER BY SUM(candidates.score) DESC
LIMIT ?%d;
`

	//query to determine ETag

	// Args: [1] code const version
	queryEtagAll = `
WITH src AS (
SELECT DISTINCT s.hash
FROM document AS d
JOIN source AS s ON s.id = d.source_id
WHERE s.status = 'completed'
ORDER BY s.id
)
SELECT
?1 || '|' ||
COALESCE(
group_concat(hash, '|'),
''
) AS state
FROM src;
`

	queryMetaDoc = `
SELECT
?1 || '|' || s.hash AS state,
s.status,
d.name
FROM document AS d
JOIN source AS s ON s.id = d.source_id
WHERE d.id = ?2
`
)

func prepareQueries(db store.Read) (queries queries, err error) {
	size := maxContextWords + 1

	queries.autocomplete.doc = make([]*sql.Stmt, size)
	queries.autocomplete.all = make([]*sql.Stmt, size)
	queries.nextWord.doc = make([]*sql.Stmt, size)
	queries.nextWord.all = make([]*sql.Stmt, size)

	if queries.eTagAll, err = db.Prepare(queryEtagAll); err != nil {
		return queries, fmt.Errorf("etag all: %w", err)
	}
	if queries.metaDoc, err = db.Prepare(queryMetaDoc); err != nil {
		return queries, fmt.Errorf("etag doc: %w", err)
	}
	if queries.autocomplete.doc[0], err = db.Prepare(queryAutocompleteDocZero); err != nil {
		return queries, fmt.Errorf("autocomplete zero doc: %w", err)
	}
	if queries.autocomplete.all[0], err = db.Prepare(queryAutocompleteAllZero); err != nil {
		return queries, fmt.Errorf("autocomplete zero all: %w", err)
	}
	if queries.nextWord.doc[0], err = db.Prepare(queryNextWordDocZero); err != nil {
		return queries, fmt.Errorf("next word zero doc: %w", err)
	}
	if queries.nextWord.all[0], err = db.Prepare(queryNextWordAllZero); err != nil {
		return queries, fmt.Errorf("next word zero all: %w", err)
	}

	for contextLength := 1; contextLength <= maxContextWords; contextLength++ {
		var b strings.Builder

		b.WriteString("WITH\n")
		for i := 1; i <= contextLength; i++ {
			if i > 1 {
				b.WriteString(",\n")
			}
			fmt.Fprintf(&b, cteContext, i)
		}

		queryAll := b.String() + ",\n" + cteStatus + queryStart
		b.WriteString(",\n")
		fmt.Fprintf(&b, cteDocumentFilter, contextLength+1)
		b.WriteString(queryStart)

		queryDoc := b.String()

		b.Reset()

		for n := contextLength + 1; n >= 2; n-- {
			if n < contextLength+1 {
				queryAll += unionAll
				queryDoc += unionAll
			}

			contextCount := n - 1
			whereOffset := contextLength - contextCount

			distanceFromPeak := float64(n) - preferredNgramLength
			weight := 1.0 / (1.0 + distanceFromPeak*distanceFromPeak)

			damping := math.Pow(lowCountDampingBase, float64(n-2))

			var whereParts []string
			for k := 1; k <= contextCount; k++ {
				whereParts = append(whereParts, fmt.Sprintf(where, k, whereOffset+k))
			}
			whereClause := strings.Join(whereParts, and) + and + whereSource

			queryAll += fmt.Sprintf(ngram, n, weight, n-1, whereClause, damping)

			queryDoc += fmt.Sprintf(ngramDoc, n, weight, whereClause, damping)
		}

		queries.autocomplete.all[contextLength], err = db.Prepare(queryAll + fmt.Sprintf(autocompleteQueryEnd, contextLength+1, contextLength+2, contextLength+3))
		if err != nil {
			return queries, fmt.Errorf("autocomplete all [%d]: %w", contextLength, err)
		}
		queries.autocomplete.doc[contextLength], err = db.Prepare(queryDoc + fmt.Sprintf(autocompleteQueryEnd, contextLength+2, contextLength+3, contextLength+4))
		if err != nil {
			return queries, fmt.Errorf("autocomplete doc [%d]: %w", contextLength, err)
		}
		queries.nextWord.all[contextLength], err = db.Prepare(queryAll + fmt.Sprintf(nextWordQueryEnd, contextLength+1))
		if err != nil {
			return queries, fmt.Errorf("next word all [%d]: %w", contextLength, err)
		}
		queries.nextWord.doc[contextLength], err = db.Prepare(queryDoc + fmt.Sprintf(nextWordQueryEnd, contextLength+2))
		if err != nil {
			return queries, fmt.Errorf("next word doc [%d]: %w", contextLength, err)
		}
	}
	return
}
