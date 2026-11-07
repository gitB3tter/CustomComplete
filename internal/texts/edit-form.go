package texts

import (
	"context"
	"customcomplete/internal/catalog"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	_ "embed"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	FileName          catalog.DocumentFilename
	Filesize          int64
	WordCount         int64
	WordCountFmt      string
	ETagStrong        transport.ETagStrong
}

func newEditFormData(userData FormUserData, doc catalog.Document, err error, ETagStrong transport.ETagStrong, langTag language.Tag) (formdata EditFormData) {
	formdata = EditFormData{
		formData:     newFormData(userData, err, langTag),
		FileName:     doc.Filename,
		Filesize:     doc.Filesize,
		WordCount:    doc.WordCount,
		WordCountFmt: message.NewPrinter(langTag).Sprintf("%d", doc.WordCount),
		ETagStrong:   ETagStrong,
	}
	if err, ok := errors.AsType[*transport.ErrorPrecondition](err); ok {
		formdata.ErrorPrecondition = err
	}
	return formdata
}

func (ed EditFormData) HasErrors() bool {
	return ed.ErrorPrecondition != nil || len(ed.Errors) > 0
}

func (h *handler) registerEdit(mux *http.ServeMux) {
	mux.Handle("GET /texts/{textID}/edit",
		transport.NoBody(
			h.editForm,
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

	docID, doc, proceed := h.getDocOrRespond(w, r, r.PathValue("textID"))
	if !proceed {
		return
	}

	var ETagStrong transport.ETagStrong
	if ETagStrong, proceed = itemEtagStrongOr500(w, r, docID, doc); !proceed {
		return
	}

	w.Header().Set("Link", itemLinkUp(docID, doc.Name).String())

	output, err := h.bufferEditForm(r.Context(), docID, newEditFormData(FormUserData{}, doc, nil, ETagStrong, langTag))
	if err != nil {
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return
	}

	transport.AddVaryOnce(w.Header(), "Cookie")
	etag := transport.GetEtagWeak(output)
	etagMatch := transport.ETagMatchWeak(r.Header.Values("If-None-Match"), etag)
	w.Header().Set("ETag", string(etag))
	if etagMatch {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", mediatypes.HTML.String())
	w.Header().Set("Content-Length", strconv.Itoa(len(output)))
	w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodHead, http.MethodOptions}, ", "))
	transport.Write(r.Context(), w, output)
}

func (h *handler) bufferEditForm(ctx context.Context, docID catalog.DocumentID, data EditFormData) ([]byte, error) {

	lang := i18n.LangTagFromContext(ctx)
	T := i18n.Func(lang)

	var buffer []byte
	var err error

	if buffer, err = ui.BufferHTML(h.editFormTemplate[lang], struct {
		Base ui.BaseDataHTML
		ID   catalog.DocumentID
		EditFormData
	}{
		Base: ui.BaseDataHTML{
			Title:           T(i18n.NavigationTextEditForm, data.Name),
			MetaDescription: T(i18n.TextEditFormMetaDescription, data.Name),
			Lang:            lang,
			HasError:        data.HasErrors(),
		},
		ID:           docID,
		EditFormData: data,
	}); err != nil {
		return nil, err
	}

	return buffer, nil

}
