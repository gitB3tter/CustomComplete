package catalog

import (
	"customcomplete/internal/constraints"
	"customcomplete/internal/errs"
	"customcomplete/internal/store"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type textStore struct {
	readWrite store.ReadWrite
	read      store.Read
}

//note: used sqlite driver seems to not (properly) implement func (sql.Result) RowsAffected() (int64, error), therefore RETURNING workaround

func (s *textStore) create(name constraints.Label, filename constraints.SafeFilename, hash constraints.Sha256Hex, filesize int64, contents []byte) (docID DocumentID, sourceID sourceID, srcStatus SourceStatus, wordcount int64, err error) {
	var tx *sql.Tx

	tx, err = s.readWrite.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = tx.QueryRow(`
	INSERT INTO source(size, hash, content)
	VALUES(?, ?, ?)
	ON CONFLICT(hash) DO UPDATE SET hash = excluded.hash
	RETURNING id, status, COALESCE(wordcount, 0);
	`, filesize, hash, contents).Scan(&sourceID, &srcStatus, &wordcount); err != nil {
		return
	}

	res, err := tx.Exec(`
	INSERT INTO document(name, filename, source_id)
	VALUES(?,?,?)`,
		name, filename, sourceID)
	if err != nil {
		return
	}

	var lastInsert int64
	lastInsert, err = res.LastInsertId()
	if err != nil {
		return
	}

	docID = DocumentID(lastInsert)

	err = tx.Commit()
	if err != nil {
		return
	}

	return
}

type docPatchQuery struct {
	queryPatch string
	args       []any
}

func newDocPatchQuery(docID DocumentID, patch DocumentPatchMetadata, newSourceID *sourceID) (patchQuery *docPatchQuery) {
	patchQuery = &docPatchQuery{}
	set := make([]string, 0, 2)
	if patch.Name != nil {
		set = append(set, "name = ?")
		patchQuery.args = append(patchQuery.args, *patch.Name)
	}
	if patch.Filename != nil {
		set = append(set, "filename = ?")
		patchQuery.args = append(patchQuery.args, *patch.Filename)
	}
	if newSourceID != nil {
		set = append(set, "source_id = ?")
		patchQuery.args = append(patchQuery.args, *newSourceID)
	}
	if len(patchQuery.args) == 0 {
		return nil
	}
	patchQuery.args = append(patchQuery.args, docID)
	patchQuery.queryPatch = fmt.Sprintf(`UPDATE document
SET %s
WHERE id = ?
RETURNING
name,
filename,
(SELECT size FROM source WHERE id = source_id),
(SELECT COALESCE(wordcount, 0) FROM source WHERE id = source_id),
(SELECT status FROM source WHERE id = source_id),
(SELECT hash FROM source WHERE id = source_id)`,
		strings.Join(set, ", "))
	return
}

type QueryRow interface {
	QueryRow(query string, args ...any) *sql.Row
}

func (query docPatchQuery) queryOn(qr QueryRow) (doc Document, err error) {
	err = qr.QueryRow(query.queryPatch, query.args...).Scan(&doc.Name, &doc.Filename, &doc.Filesize, &doc.WordCount, &doc.Status, &doc.Hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return doc, new(errs.NotFound())
		}
		return doc, fmt.Errorf("patch document metadata: %w", err)
	}
	return

}

func (s textStore) updateDocument(docID DocumentID, patchMeta DocumentPatchMetadata, contents SourceContents, hash constraints.Sha256Hex) (doc Document, srcID sourceID, err error) {
	fail := func(err error) (Document, sourceID, error) {
		return doc, srcID, err
	}

	tx, err := s.readWrite.Begin()
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	var size int64
	var wordcount int64
	var status SourceStatus

	if err = tx.QueryRow(`
INSERT INTO source(size, hash, content)
VALUES(?, ?, ?)
ON CONFLICT(hash) DO UPDATE SET hash = excluded.hash
RETURNING id, size, COALESCE(wordcount, 0), status;
	`, len(contents), hash, contents).Scan(&srcID, &size, &wordcount, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fail(new(errs.NotFound()))
		}
		return fail(fmt.Errorf("update source: %w", err))
	}

	patchQuery := newDocPatchQuery(docID, patchMeta, &srcID)

	if doc, err = patchQuery.queryOn(tx); err != nil {
		return fail(err)
	}

	if err = tx.Commit(); err != nil {
		return fail(err)
	}
	doc.Filesize = size
	doc.WordCount = wordcount
	doc.Status = status
	return
}

func (s textStore) patchDocumentMetadata(docID DocumentID, patch DocumentPatchMetadata) (doc Document, err error) {

	patchQuery := newDocPatchQuery(docID, patch, nil)

	return patchQuery.queryOn(s.readWrite)
}

func (s *textStore) getDocument(docID DocumentID) (Document, error) {
	var doc Document

	if err := s.read.QueryRow(`
SELECT
d.name, d.filename, src.size, COALESCE(src.wordcount, 0), src.status, src.hash
FROM document d
INNER JOIN source src ON src.id=d.source_id
WHERE d.id=?`, docID).Scan(&doc.Name, &doc.Filename, &doc.Filesize, &doc.WordCount, &doc.Status, &doc.Hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Document{}, new(errs.NotFound())
		}
		return Document{}, err
	}
	return doc, nil
}

func (s *textStore) getDocumentRefs() ([]DocumentRef, error) {
	rows, err := s.read.Query(`
SELECT
d.id, d.name
FROM document d
INNER JOIN source s ON d.source_id=s.id
ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var docRefs []DocumentRef

	for rows.Next() {
		var docRef DocumentRef
		if err = rows.Scan(&docRef.ID, &docRef.Name); err != nil {
			return nil, err
		}
		docRefs = append(docRefs, docRef)
	}
	err = rows.Err()

	if err != nil {
		return nil, err
	}

	return docRefs, nil
}

func (s *textStore) deleteDocument(docID DocumentID) error {

	var rawName string
	err := s.readWrite.QueryRow(`
DELETE FROM document
WHERE id = ?
RETURNING name`, docID).Scan(&rawName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return new(errs.NotFound())
		}
		return fmt.Errorf("failed deleting document %d: %w", docID, err)
	}
	if rawName == "" {
		return new(errs.NotFound())
	}

	return nil
}

func (s *textStore) getSourceID(docID DocumentID) (sourceID, error) {
	var sourceId int64
	if err := s.read.QueryRow(`
SELECT
s.id
FROM document d
INNER JOIN source s ON d.source_id=s.id
WHERE d.id = ?`, docID).Scan(&sourceId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, new(errs.NotFound())
		}
		return 0, err
	}
	return sourceID(sourceId), nil
}

func (s *textStore) getContents(docID DocumentID) (content SourceContents, err error) {
	if err = s.read.QueryRow(`SELECT s.content
	FROM source s
	JOIN document d on d.source_id=s.id
	WHERE d.id=?`, docID).Scan(&content); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, new(errs.NotFound())
		}
		return nil, err
	}
	return content, nil
}
