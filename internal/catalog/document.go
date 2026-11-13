package catalog

import (
	"customcomplete/internal/constraints"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/progress"
	"errors"
	"log/slog"
	"strconv"
)

const (
	DocumentNameMaxCodePoints = constraints.ShortStringMaxCodePoints
)

// > 0 (maybe put also as general/global/primitive constraint?)
type DocumentID int64

type Document struct {
	Name      constraints.Label
	Filename  constraints.SafeFilename
	Filesize  int64
	WordCount int64
	Status    SourceStatus
	Hash      constraints.Sha256Hex
}

// id and label
type DocumentRef struct {
	ID   DocumentID
	Name constraints.Label
}

func (service *Service) CreateDocument(name constraints.Label, filename constraints.SafeFilename, contents SourceContents) (doc Document, id DocumentID, err error) {

	var srcID sourceID

	doc.Name = name
	doc.Filesize = int64(len(contents))
	doc.Hash = constraints.NewSha256Hex(contents)
	doc.Filename = filename

	id, srcID, doc.Status, doc.WordCount, err = service.store.create(name, doc.Filename, doc.Hash, doc.Filesize, contents)
	if err != nil {
		return Document{}, 0, err
	}

	if doc.Status == StatusPending {
		service.broker.Update(int64(srcID), 0, i18n.TextStatusPending)
		err := service.enqueue(srcID)
		if err != nil && !errors.Is(err, ErrQueueFull) {
			slog.Warn("enqueue Job source processing: " + err.Error())
		}
	}

	return doc, id, nil
}

type DocumentPatchMetadata struct {
	Name     *constraints.Label
	Filename *constraints.SafeFilename
}

func (d DocumentPatchMetadata) Empty() bool {
	empty := DocumentPatchMetadata{}
	return d == empty
}

func (service *Service) UpdateDocument(docID DocumentID, patch DocumentPatchMetadata, contents SourceContents) (doc Document, err error) {

	var srcID sourceID

	if doc, srcID, err = service.store.updateDocument(docID, patch, contents, constraints.NewSha256Hex(contents)); err != nil {
		return
	}

	if doc.Status == StatusPending {
		service.broker.Update(int64(srcID), 0, i18n.TextStatusPending)
		err := service.enqueue(srcID)
		if err != nil && !errors.Is(err, ErrQueueFull) {
			slog.Warn("enqueue Job source processing: " + err.Error())
		}
	}

	return
}

func (service *Service) PatchDocumentMetadata(id DocumentID, patch DocumentPatchMetadata) (doc Document, err error) {
	if patch.Empty() {
		return service.GetDocument(id)
	}
	return service.store.patchDocumentMetadata(id, patch)
}

func (service *Service) GetDocumentRefs() ([]DocumentRef, error) {
	return service.store.getDocumentRefs()
}

func (service *Service) GetDocumentStatus(id DocumentID) (SourceStatus, error) {
	doc, err := service.store.getDocument(id)
	return doc.Status, err
}

func (service *Service) GetDocumentProgressLatest(id DocumentID) (progress.Progress, error) {
	sourceID, err := service.store.getSourceID(id)
	if err != nil {
		return progress.Progress{}, err
	}
	latest, ok := service.broker.GetLatest(int64(sourceID))
	if !ok {
		return progress.Progress{}, new(errs.NotFound())
	}
	return latest, nil
}

func (service *Service) SubscribeDocumentProgress(id DocumentID) (updates <-chan progress.Progress, latest progress.Progress, hasLatest bool, unsubscribe func(), err error) {
	sourceID, err := service.store.getSourceID(id)
	if err != nil {
		return nil, progress.Progress{}, false, nil, err
	}
	updates, latest, hasLatest, unsubscribe = service.broker.Subscribe(int64(sourceID))
	return
}

func (service *Service) GetDocument(id DocumentID) (Document, error) {
	doc, err := service.store.getDocument(id)
	return doc, err
}

func (service *Service) DeleteDocument(id DocumentID) error {
	return service.store.deleteDocument(id)
}

func ParseDocumentIDFromInt64(numb int64) (DocumentID, bool) {
	if numb <= 0 {
		return 0, false
	}
	return DocumentID(numb), true
}

func ParseDocumentIDFromString(str string) (DocumentID, bool) {
	numb, err := strconv.ParseInt(str, 10, 64)
	if err != nil || numb <= 0 {
		return 0, false
	}
	return DocumentID(numb), true
}
