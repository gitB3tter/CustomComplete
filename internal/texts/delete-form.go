package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/constraints"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"net/http"
)

var (
	//go:embed partials/delete-form.html
	deleteFormPartialTemplate string
	//go:embed delete-form.html
	deleteFormTemplate string
)

type deleteFormData struct {
	ID         catalog.DocumentID
	Name       constraints.Label
	ETagStrong transport.ETagStrong
}

func DeleteFormPath(docID catalog.DocumentID) weblinking.Path {
	return TextsItemPath(docID).Join("delete")
}

func (h *handler) registerDeleteForm(mux *http.ServeMux) {
	mux.Handle("GET /texts/{textID}/delete",
		transport.NoBody(h.deleteForm),
	)
	mux.Handle("OPTIONS /texts/{textID}/delete",
		transport.NoBody(h.deleteFormOptions),
	)
}

func (h *handler) initDeleteForm(base *ui.BaseTemplates) (err error) {
	h.deleteFormTemplate, err = base.BuildHTML(deleteFormPartialTemplate, deleteFormTemplate)
	return
}

func (h *handler) deleteForm(w http.ResponseWriter, r *http.Request) {

	langTag := i18n.LangTagFromContext(r.Context())

	if proceed := transport.RequireAccepts(mediatypes.HTML, w, r, collectionLink(langTag)); !proceed {
		return
	}

	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}

	var output []byte
	var err error
	T := i18n.TranslateFuncFromContext(r.Context())
	output, err = ui.BufferHTML(h.deleteFormTemplate[langTag],
		struct {
			Base ui.BaseDataHTML
			deleteFormData
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.NavigationTextDeleteForm, doc.Name),
				MetaDescription: T(i18n.DocsRelsDeleteFormMetaDescription),
				NoIndex:         true,
			},
			deleteFormData: deleteFormData{
				ID:         docID,
				Name:       doc.Name,
				ETagStrong: itemETagStrong(doc),
			},
		})
	if err != nil {
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return
	}

	w.Header().Set("Link", itemLinkUp(docID, doc.Name).String())
	transport.SetNoCache(w.Header())
	etag := transport.GetEtagWeak(output)
	w.Header().Set("ETag", string(etag))
	if transport.ETagMatchWeak(r.Header.Values("If-None-Match"), etag) {
		transport.WriteNoBody(r.Context(), w, http.StatusNotModified)
		return
	}

	transport.Write(r.Context(), w, output, mediatypes.HTML, http.StatusOK)

}

func (h *handler) deleteFormOptions(w http.ResponseWriter, r *http.Request) {
	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}

	w.Header().Set("Allow", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Link, Allow, Vary")
	w.Header().Set("Accept-Post", mediatypes.FormUrlEncoded.String())
	w.Header().Set("Link", itemLinkUp(docID, doc.Name).String())
	transport.WriteNoBody(r.Context(), w, http.StatusNoContent)
}
