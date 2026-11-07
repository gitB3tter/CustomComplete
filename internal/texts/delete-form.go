package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	_ "embed"
	"errors"
	"net/http"
)

var (
	//go:embed partials/delete-form.html
	templatePartialDeleteForm string
	//go:embed delete-form.html
	templateDeleteForm string
)

type deleteFormData struct {
	ID         catalog.DocumentID
	Name       catalog.DocumentName
	ETagStrong transport.ETagStrong
}

func (h *handler) registerConfirmDelete(mux *http.ServeMux) {
	mux.Handle("GET /texts/{textID}/delete",
		transport.NoBody(h.deleteForm))
}

func (h *handler) initDeleteForm(base *ui.BaseTemplates) (err error) {
	h.deleteFormTemplate, err = base.BuildHTML(templatePartialDeleteForm, templateDeleteForm)
	return
}

func (h *handler) deleteForm(w http.ResponseWriter, r *http.Request) {

	if ok := transport.RequireAccepts(mediatypes.HTML, w, r, collectionLink(i18n.LangTagFromContext(r.Context()))); !ok {
		return
	}

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
	if ETagStrong, ok = itemEtagStrongOr500(w, r, id, doc); !ok {
		return
	}

	var output []byte

	w.Header().Set("Content-Type", mediatypes.HTML.String())
	T := i18n.TranslateFuncFromContext(r.Context())
	output, err = ui.BufferHTML(h.deleteFormTemplate[i18n.LangTagFromContext(r.Context())],
		struct {
			Base ui.BaseDataHTML
			deleteFormData
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.TextDeleteFormTitle, doc.Name),
				MetaDescription: T(i18n.DocsRelsDeleteFormMetaDescription),
				NoIndex:         true,
			},
			deleteFormData: deleteFormData{
				ID:         id,
				Name:       doc.Name,
				ETagStrong: ETagStrong,
			},
		})
	if err != nil {
		respondInternalErrorSubresource(w, r, err, id, doc.Name)
		return
	}
	w.Header().Set("Link", itemLinkUp(id, doc.Name).String())
	transport.SetNoCache(w.Header())
	etag := transport.GetEtagWeak(output)
	if transport.ETagMatchWeak(r.Header.Values("If-None-Match"), etag) {
		transport.RespondNoContent(w)
		return
	}

	transport.Write(r.Context(), w, output)

}
