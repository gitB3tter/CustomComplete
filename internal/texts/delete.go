package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"errors"
	"net/http"
	"net/url"
)

func (h *handler) registerDelete(mux *http.ServeMux) {
	mux.Handle("POST /texts/{textID}/delete",
		transport.LimitBody(h.delete, mediatypes.Types{mediatypes.URLEncoded}, 256))
	mux.Handle("DELETE /texts/{textID}",
		transport.NoBody(h.delete))
}

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := catalog.ParseDocumentIDFromString(r.PathValue("textID"))
	if !ok {
		respondNotFound(w, r)
		return
	}

	doc, err := h.service.GetDocument(id)
	if err != nil {
		var errWithCode *errs.CodedError
		if errors.As(err, &errWithCode) && errWithCode.Code() == errs.CodeNotFound {
			respondNotFound(w, r)
		} else {
			respondInternalError(w, r, err)
		}
		return
	}

	var ETagStrong transport.ETagStrong
	var proceed bool
	if ETagStrong, proceed = itemEtagStrongOr500(w, r, id, doc); !proceed {
		return
	}

	langTag := i18n.LangTagFromContext(r.Context())

	var errPreconditon *transport.ErrorPrecondition

	if errPreconditon = transport.Precondition(ETagStrong, r.FormValue(ui.HiddenFieldEtag), langTag); errPreconditon != nil {
		errPreconditon = transport.Precondition(ETagStrong, r.Header.Get("If-Match"), langTag)
	}

	if errPreconditon != nil {
		h.failPrecondition(w, r, id, doc, *errPreconditon, ETagStrong)
		return
	}

	if err = h.service.DeleteDocument(id); err == nil {
		if transport.IsFullDocumentNavigation(r) {
			http.SetCookie(w, &http.Cookie{
				Name:   string(flashDeleted),
				Value:  url.QueryEscape(string(doc.Name)),
				MaxAge: 60,
				Path:   collectionPath().String(),
			})
			transport.AddVaryOnce(w.Header(), "Accept")
			transport.SeeOther(w, r, collectionPath(), i18n.Func(langTag)(i18n.TextDeleteFlashSuccess, doc.Name))
			return
		}
		w.Header().Set("Link", collectionLink(langTag).String())
		w.Header().Del("Vary")
		w.Header().Del("Content-Language")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var errWithCode *errs.CodedError
	if errors.As(err, &errWithCode) && errWithCode.Code() == errs.CodeNotFound {
		respondNotFound(w, r)
		return
	}

	respondInternalErrorSubresource(w, r, err, id, doc.Name)
}

func (h *handler) failPrecondition(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, doc catalog.Document, errPrecondition transport.ErrorPrecondition, ETagStrong transport.ETagStrong) {

	T := i18n.TranslateFuncFromContext(r.Context())

	if transport.IsFullDocumentNavigation(r) {
		http.SetCookie(w, &http.Cookie{
			Name:   errPrecondition.CustomMessage(string(flashDeletePreconditionFailed), string(flashDeleteNoPrecondition)),
			MaxAge: 60,
			Path:   itemPath(docID).String(),
		})
		transport.SeeOther(w, r, itemPath(docID), errPrecondition.CustomMessage(T(i18n.TextDeletePrecondition), T(i18n.TextDeleteNoPrecondition)))
		return
	}

	errResponseType := transport.ResolveErrorContentType(r)
	w.Header().Set("Content-Type", errResponseType.String())

	if errResponseType == mediatypes.HTML {
		h.respondItemPreconditionHTML(w, r, docID, doc, &errPrecondition, ETagStrong)
		return
	}
	var output []byte
	switch errResponseType {
	case mediatypes.JSON, mediatypes.ProblemJSON:
		var err error
		output, err = transport.ProblemGeneric(errPrecondition.Error())
		if err != nil {
			respondInternalErrorSubresource(w, r, err, docID, doc.Name)
			return
		}
	case mediatypes.TextPlain:
		output = []byte(errPrecondition.Error())
	}

	w.WriteHeader(int(errPrecondition.Status()))
	transport.Write(r.Context(), w, output)
}
