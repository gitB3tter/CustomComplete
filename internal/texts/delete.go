package texts

import (
	"cmp"
	"customcomplete/internal/catalog"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"net/http"
	"net/url"
)

func (h *handler) registerDelete(mux *http.ServeMux) {
	mux.Handle("POST /texts/{textID}/delete",
		transport.LimitBody(
			h.delete,
			mediatypes.Types{mediatypes.URLEncoded},
			256,
		),
	)
	mux.Handle("DELETE /texts/{textID}",
		transport.NoBody(h.delete),
	)
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {

	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}

	langTag := i18n.LangTagFromContext(r.Context())

	etagStrong := itemETagStrong(doc)
	if errPrecondition := transport.Precondition(etagStrong, cmp.Or(r.FormValue(ui.HiddenFieldEtag), r.Header.Get("If-Match")), langTag); errPrecondition != nil {
		h.respondDeleteErrors(w, r, docID, doc, *errPrecondition, etagStrong)
		return
	}

	if err := h.service.DeleteDocument(docID); err != nil {
		if errs.HasCode(err, errs.CodeNotFound) {
			respondNotFound(w, r)
			return
		}
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return
	}

	if transport.IsFullDocumentNavigation(r) {
		http.SetCookie(w, &http.Cookie{
			Name:   string(flashDeleted),
			Value:  url.QueryEscape(string(doc.Name)),
			MaxAge: 60,
			Path:   CollectionPath.URL().Path,
		})
		transport.AddVaryOnce(w.Header(), "Accept")
		transport.SeeOther(w, r, CollectionPath, i18n.Func(langTag)(i18n.TextDeleteFlashSuccess, doc.Name))
		return
	}
	w.Header().Set("Link", collectionLink(langTag).String())
	w.Header().Del("Vary")
	w.Header().Del("Content-Language")
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) respondDeleteErrors(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, doc catalog.Document, errPrecondition transport.ErrorPrecondition, etagStrong transport.ETagStrong) {

	langTag := i18n.LangTagFromContext(r.Context())
	T := i18n.Func(langTag)

	if transport.IsFullDocumentNavigation(r) {
		http.SetCookie(w, &http.Cookie{
			Name:   errPrecondition.CustomMessage(string(flashDeletePreconditionFailed), string(flashDeleteNoPrecondition)),
			MaxAge: 60,
			Path:   TextsItemPath(docID).String(),
		})
		transport.SeeOther(w, r, TextsItemPath(docID), errPrecondition.CustomMessage(T(i18n.TextDeletePrecondition), T(i18n.TextDeleteNoPrecondition)))
		return
	}

	responseType := transport.ResolveErrorContentType(r)
	w.Header().Set("Content-Type", responseType.String())
	transport.SetErrorHeaders(w.Header())

	var err error
	var output []byte
	switch responseType {
	case mediatypes.HTML:
		output, err = h.bufferItemHTML(w, r, doc, docID, etagStrong, langTag, &errPrecondition)
		if err != nil {
			respondInternalErrorSubresource(w, r, err, docID, doc.Name)
			return
		}
		writeItem(r.Context(), w, docID, doc, output, errPrecondition.Status(), langTag, mediatypes.HTML)
		return
	case mediatypes.JSON, mediatypes.ProblemJSON:
		output, err = transport.ProblemGeneric(errPrecondition.Error())
		if err != nil {
			respondInternalErrorSubresource(w, r, err, docID, doc.Name)
			return
		}
	case mediatypes.TextPlain:
		output = []byte(errPrecondition.Error())
	}

	transport.Write(r.Context(), w, output, responseType, int(errPrecondition.Status()))
}
