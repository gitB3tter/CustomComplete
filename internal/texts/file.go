package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/errs"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/weblinking"
	"errors"
	"net/http"
)

func TextsFilePath(docID catalog.DocumentID) weblinking.Path {
	return TextsItemPath(docID).Join("file")
}

func (h *handler) registerFile(mux *http.ServeMux) {
	mux.Handle("GET /texts/{textID}/file",
		transport.NoBody(h.file))
}

func (h *handler) file(w http.ResponseWriter, r *http.Request) {

	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}
	if !transport.RequireAccepts(mediatypes.TextPlain, w, r, itemLinkUp(docID, doc.Name)) {
		return
	}

	links := weblinking.Links{
		itemLinkUp(docID, doc.Name),
		weblinking.New(weblinking.Self, TextsFilePath(docID)),
	}

	etagStrong := transport.GetEtagStrongLiteral(string(doc.Hash))
	if transport.EtagMatchStrong(r.Header.Values("If-None-Match"), etagStrong) {
		w.Header().Set("Link", links.String())
		w.WriteHeader(http.StatusNotModified)
		return
	}

	var err error
	var contents catalog.SourceContents
	if contents, err = h.service.GetSourceContents(docID); err != nil {
		var errWithCode *errs.CodedError
		if errors.As(err, &errWithCode) && errWithCode.Code() == errs.CodeNotFound {
			respondNotFound(w, r)
		} else {
			respondInternalError(w, r, err)
		}
		return
	}

	if r.URL.Query().Has("download") {
		w.Header().Set("Content-Disposition", `attachment; filename="`+string(doc.Filename)+`.txt"`)
	} else {
		w.Header().Set("Content-Disposition", `inline; filename="`+string(doc.Filename)+`.txt"`)
	}

	w.Header().Set("ETag", string(etagStrong))

	w.Header().Set("Link", links.String())

	transport.Write(r.Context(), w, contents, mediatypes.TextPlain, http.StatusOK)
}
