package texts

import (
	"context"
	"customcomplete/internal/catalog"
	"customcomplete/internal/docs"
	"customcomplete/internal/errs"
	"customcomplete/internal/home"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	supportedTypesItem = mediatypes.Types{mediatypes.HTML, mediatypes.JSON}
	//go:embed item.html
	templateItem string
)

type itemData struct {
	Name     string `json:"name"`
	FileName string `json:"filename"`
	FileSize int64  `json:"filesize"`
	FileHash string `json:"filehash"`
}

func newItemData(doc catalog.Document) itemData {
	return itemData{
		Name:     string(doc.Name),
		FileName: string(doc.Filename),
		FileSize: doc.Filesize,
		FileHash: string(doc.Hash),
	}
}

func getStatusTranslation(status catalog.SourceStatus, lang language.Tag) string {
	var message i18n.Key
	switch status {
	case catalog.StatusPending:
		message = i18n.TextStatusPending
	case catalog.StatusProcessing:
		message = i18n.TextStatusProcessing
	case catalog.StatusFailed:
		message = i18n.TextStatusFailed
	case catalog.StatusComplete:
		message = i18n.TextStatusComplete
	}
	return i18n.Translate(lang, message)
}

func (h *handler) registerDetail(mux *http.ServeMux) {
	mux.Handle("GET /texts/{id}",
		transport.NoBody(
			h.item,
		))
}

func (h *handler) initItem(base *ui.BaseTemplates) (err error) {
	h.itemTemplate, err = base.BuildHTML(templatePartialDeleteForm, templateStatusPartialsHTML, templateItem)
	return
}

func (h *handler) item(w http.ResponseWriter, r *http.Request) {

	responseType, ok := transport.NegotiateOr406(w, r, collectionLink(i18n.LangTagFromContext(r.Context())), supportedTypesItem)
	if !ok {
		return
	}

	docID, ok := catalog.ParseDocumentIDFromString(r.PathValue("id"))
	if !ok {
		respondNotFound(w, r)
		return
	}

	doc, err := h.service.GetDocument(docID)
	if err != nil {
		var errWithCode *errs.CodedError
		if errors.As(err, &errWithCode) && errWithCode.Code() == errs.CodeNotFound {
			respondNotFound(w, r)
		} else {
			respondInternalError(w, r, err)
		}
		return
	}
	ETagStrong, err := itemETagStrong(doc)
	if err != nil {
		respondInternalError(w, r, err)
		return
	}
	setItemHeaders(w.Header())

	langTag := i18n.LangTagFromContext(r.Context())

	if responseType == mediatypes.JSON && transport.EtagMatchStrong(r.Header.Values("If-None-Match"), ETagStrong) {
		respondItemNotModified(w, docID)
		return
	}

	var output []byte
	switch responseType {
	case mediatypes.HTML:
		var errPrecondition *transport.ErrorPrecondition
		if noPre := consumeFlash(w, r, itemPath(docID), flashDeleteNoPrecondition); noPre != "" {
			errPrecondition = new(transport.ErrPrecondition(http.StatusPreconditionRequired, langTag))
		} else if preFailed := consumeFlash(w, r, itemPath(docID), flashDeletePreconditionFailed); preFailed != "" {
			errPrecondition = new(transport.ErrPrecondition(http.StatusPreconditionFailed, langTag))
		}
		if errPrecondition != nil {
			h.respondItemPreconditionHTML(w, r, docID, doc, errPrecondition, ETagStrong)
			return
		}
		output, err = h.bufferItemHTML(w, r, doc, docID, ETagStrong, i18n.LangTagFromContext(r.Context()), errPrecondition)
	case mediatypes.JSON:
		output, err = bufferItemJSON(doc)
	}
	if err != nil {
		respondInternalError(w, r, err)
		return
	}

	writeItem(r.Context(), w, docID, doc, output, http.StatusOK, langTag, responseType)
}

func setItemHeaders(h http.Header) {
	h.Set("Allow", strings.Join([]string{http.MethodGet, http.MethodHead, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions}, ", "))
	h.Set("Accept-Patch", mediatypes.MergePatchJSON.String())
	h.Set("Accept-Post", mediatypes.MultipartForm.String())
}

func (h *handler) respondItemPreconditionHTML(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, doc catalog.Document, errPrecondition *transport.ErrorPrecondition, ETagStrong transport.ETagStrong) {

	var output []byte
	var err error

	langTag := i18n.LangTagFromContext(r.Context())

	output, err = h.bufferItemHTML(w, r, doc, docID, ETagStrong, langTag, errPrecondition)
	if err != nil {
		w.Header().Del("Location")
		respondInternalError(w, r, err)
	}

	transport.SetErrorHeaders(w.Header())

	writeItem(r.Context(), w, docID, doc, output, http.StatusConflict, langTag, mediatypes.HTML)
}

func respondItemNotModified(w http.ResponseWriter, docID catalog.DocumentID) {
	w.Header().Set("Link", append(weblinking.AllAlternates(itemPath(docID), supportedTypesItem...), itemLinks(docID)...).String())
	w.WriteHeader(http.StatusNotModified)
}

func writeItem(ctx context.Context, w http.ResponseWriter, docID catalog.DocumentID, doc catalog.Document, output []byte, statuscode int, langTag language.Tag, responsetype mediatypes.Type) {

	links := append(weblinking.Alternates(itemPath(docID), langTag, responsetype, supportedTypesItem...), itemLinks(docID)...)
	if responsetype != mediatypes.HTML {
		links = append(links, itemLinkJSONDescribedBy()...)
	}
	if doc.Status == catalog.StatusComplete {
		T := i18n.Func(langTag)
		if url, err := url.Parse("/suggestions?textID=" + strconv.FormatInt(int64(docID), 10)); err != nil {
			links = append(links, weblinking.New(weblinking.Search, url).WithTitle(T(i18n.SuggestionsTitleFilteredTextID, doc.Name)))

		}
	}
	w.Header().Set("Link", links.String())
	w.Header().Set("Content-Type", responsetype.String())
	w.Header().Set("Content-Length", strconv.Itoa(len(output)))
	w.WriteHeader(statuscode)
	transport.Write(ctx, w, output)

}

func (h *handler) bufferItemHTML(w http.ResponseWriter, r *http.Request, doc catalog.Document, docID catalog.DocumentID, ETagStrong transport.ETagStrong, langTag language.Tag, errPrecondition *transport.ErrorPrecondition) ([]byte, error) {
	T := i18n.Func(langTag)

	return ui.BufferHTML(h.itemTemplate[langTag], struct {
		Base ui.BaseDataHTML
		itemData
		Status            StatusData
		ErrorPrecondition *transport.ErrorPrecondition
		FileSizeFmt       string
		WordCountFmt      string
		ETagStrong        transport.ETagStrong
		Flash             string
		ID                catalog.DocumentID
	}{
		Base: ui.BaseDataHTML{
			Title:           string(doc.Name),
			MetaDescription: T(i18n.TextItemMetaDescription, doc.Name),
			Lang:            langTag,
			HasError:        errPrecondition != nil,
			Breadcrumbs: weblinking.Links{
				weblinking.Breadcrumb(home.Path(), T(i18n.NavigationHome)),
				collectionLink(langTag),
			},
			Links: itemLinks(docID),
		},
		itemData:          newItemData(doc),
		Status:            newStatusData(doc.Status, getStatusTranslation(doc.Status, langTag)),
		ErrorPrecondition: errPrecondition,
		FileSizeFmt:       i18n.FormatBytesLocale(doc.Filesize, langTag),
		WordCountFmt:      getWordcountFmt(doc.WordCount, langTag),
		ETagStrong:        ETagStrong,
		Flash:             consumeFlash(w, r, itemPath(docID), flashCreated, flashUpdated),
		ID:                docID,
	})
}

func bufferItemJSON(doc catalog.Document) ([]byte, error) {
	return json.Marshal(newItemData(doc))
}

func getWordcountFmt(wordcount int64, langTag language.Tag) string {
	if wordcount > 0 {
		return message.NewPrinter(langTag).Sprintf("%d", wordcount)
	} else {
		return ""
	}
}

func itemLinks(docID catalog.DocumentID) weblinking.Links {
	selfPath := itemPath(docID)
	return weblinking.Links{
		weblinking.New(weblinking.Up, collectionPath()),
		weblinking.New(weblinking.Collection, collectionPath()),
		weblinking.New(weblinking.Self, selfPath),
		weblinking.New(weblinking.Enclosure, selfPath.JoinPath("file")).WithType(mediatypes.TextPlain),
		weblinking.New(weblinking.EditMedia, selfPath.JoinPath("file")).WithType(mediatypes.TextPlain),
		weblinking.New(weblinking.Status, selfPath.JoinPath("status")),
		weblinking.New(weblinking.EditForm, selfPath.JoinPath("edit")).WithType(mediatypes.HTML),
		weblinking.New(weblinking.DeleteForm, selfPath.JoinPath("delete")).WithType(mediatypes.HTML),
	}
}

func itemLinkJSONDescribedBy() weblinking.Links {
	return weblinking.Links{weblinking.New(weblinking.DescribedBy, docs.GetPath(docs.SchemaTextsItem)).WithType(mediatypes.SchemaJSON)}
}

func respondMinimalItem(w http.ResponseWriter, ctx context.Context, doc catalog.Document, docID catalog.DocumentID, status int) {
	ETagStrong, err := itemETagStrong(doc)
	if err == nil {
		w.Header().Set("ETag", string(ETagStrong))
	} else {
		slog.WarnContext(ctx, "Could not generate Etag for updated Text")
	}
	w.Header().Set("Link", append(itemLinks(docID), weblinking.AllAlternates(itemPath(docID), supportedTypesItem...)...).String())
	w.Header().Del("Content-Type")
	w.Header().Del("Content-Language")
	w.WriteHeader(status)
}
