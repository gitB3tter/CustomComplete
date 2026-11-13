package texts

import (
	"context"
	"customcomplete/internal/catalog"
	"customcomplete/internal/constraints"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"errors"
	"net/http"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	//go:embed edit-form.html
	tmplEditForm string
)

type EditFormData struct {
	formData
	ErrorPrecondition *transport.ErrorPrecondition
	FileName          constraints.SafeFilename
	FileSizeFmt       string
	WordCount         int64
	WordCountFmt      string
	ETagStrong        transport.ETagStrong
}

func newEditFormData(userData FormUserData, doc catalog.Document, err error, etagStrong transport.ETagStrong, langTag language.Tag) (formdata EditFormData) {
	formdata = EditFormData{
		formData:     newFormData(userData, err, langTag),
		FileName:     doc.Filename,
		FileSizeFmt:  message.NewPrinter(langTag).Sprintf("%d", doc.Filesize),
		WordCount:    doc.WordCount,
		WordCountFmt: message.NewPrinter(langTag).Sprintf("%d", doc.WordCount),
		ETagStrong:   etagStrong,
	}
	if err, ok := errors.AsType[*transport.ErrorPrecondition](err); ok {
		formdata.ErrorPrecondition = err
	}
	return formdata
}

func (ed EditFormData) HasErrors() bool {
	return ed.ErrorPrecondition != nil || len(ed.Errors) > 0
}

func TextsEditFormPath(docID catalog.DocumentID) weblinking.Path {
	return TextsItemPath(docID).Join("edit")
}

func (h *handler) registerEditForm(mux *http.ServeMux) {
	mux.Handle("GET /texts/{textID}/edit",
		transport.NoBody(
			h.editForm,
		))
	mux.Handle("OPTIONS /texts/{textID}/edit",
		transport.NoBody(
			h.editFormOptions,
		))
}

func (h *handler) initEditForm(base *ui.BaseTemplates) (err error) {
	h.editFormTemplate, err = base.BuildHTML(tmplEditForm)
	return
}

func (h *handler) editForm(w http.ResponseWriter, r *http.Request) {

	langTag := i18n.LangTagFromContext(r.Context())

	if !transport.RequireAccepts(mediatypes.HTML, w, r, collectionLink(langTag)) {
		return
	}

	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}

	output, err := h.bufferEditForm(r.Context(), docID, newEditFormData(
		FormUserData{
			Name:   string(doc.Name),
			Source: keep,
		}, doc, nil, itemETagStrong(doc), langTag))
	if err != nil {
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return
	}

	transport.AddVaryOnce(w.Header(), "Cookie")
	w.Header().Set("Link", weblinking.Links{
		weblinking.New(weblinking.Self, TextsEditFormPath(docID)).
			WithTitle(i18n.Func(langTag)(i18n.NavigationTextEditForm, doc.Name)),
		itemLinkUp(docID, doc.Name),
	}.String())
	etag := transport.GetEtagWeak(output)
	w.Header().Set("ETag", string(etag))
	if transport.ETagMatchWeak(r.Header.Values("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Allow", "GET, HEAD, OPTIONS")
	transport.Write(r.Context(), w, output, mediatypes.HTML, http.StatusOK)
}

func (h *handler) editFormOptions(w http.ResponseWriter, r *http.Request) {
	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}

	w.Header().Set("Allow", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Link, Allow, Vary")
	w.Header().Del("Content-Language")
	w.Header().Set("Link", itemLinkUp(docID, doc.Name).String())
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) bufferEditForm(ctx context.Context, docID catalog.DocumentID, data EditFormData) ([]byte, error) {

	langTag := i18n.LangTagFromContext(ctx)
	T := i18n.Func(langTag)

	if output, err := ui.BufferHTML(h.editFormTemplate[langTag], struct {
		Base ui.BaseDataHTML
		ID   catalog.DocumentID
		EditFormData
	}{
		Base: ui.BaseDataHTML{
			Title:           T(i18n.NavigationTextEditForm, data.Name),
			MetaDescription: T(i18n.TextEditFormMetaDescription, data.Name),
			Lang:            langTag,
			HasError:        data.HasErrors(),
		},
		ID:           docID,
		EditFormData: data,
	}); err != nil {
		return nil, err
	} else {
		return output, nil
	}
}
