package texts

import (
	"context"
	"customcomplete/internal/catalog"
	"customcomplete/internal/constraints"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/schema"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	supportedTypesItem = mediatypes.Types{mediatypes.HTML, mediatypes.JSON}
	//go:embed item.html
	templateItem   string
	ItemDocPath    weblinking.Path = docPath.Join("item")
	ItemSpecPath   weblinking.Path = specPath.Join("item")
	ItemSchemaPath weblinking.Path = schemaPath.Join("item")
)

type itemData struct {
	Name     constraints.Label        `json:"name"`
	FileName constraints.SafeFilename `json:"filename"`
	FileSize int64                    `json:"filesize"`
	FileHash constraints.Sha256Hex    `json:"filehash"`
}

func newItemData(doc catalog.Document) itemData {
	return itemData{
		Name:     doc.Name,
		FileName: doc.Filename,
		FileSize: doc.Filesize,
		FileHash: doc.Hash,
	}
}

func ItemSchema(langTag language.Tag) schema.Root {
	return itemData{}.schema(langTag)
}

func (i itemData) schema(langTag language.Tag) schema.Root {
	T := i18n.Func(langTag)
	filecontentsSchema := new(catalog.SourceContents).Schema()
	return schema.NewRoot(
		ItemSchemaPath,
		map[string]schema.Schema{
			nameField: i.Name.Schema().
				WithMeta(
					T(i18n.TextFieldName),
					T(i18n.TextFieldNameHint, catalog.DocumentNameMaxCodePoints),
				),
			filenameField: i.FileName.Schema(),
			"filesize": schema.From(i.FileSize).
				WithMaximum(filecontentsSchema.XConstraints.MaxBytes).
				WithMinimum(filecontentsSchema.XConstraints.MinBytes).
				AsReadOnly(),
			"filehash": i.FileHash.Schema().
				AsReadOnly(),
		},
	).WithTitle(T(i18n.NavigationDocsTextsItem))
}

func TextsItemPath(docID catalog.DocumentID) weblinking.Path {
	return CollectionPath.Join(strconv.FormatInt(int64(docID), 10))
}

func itemLinkUp(docID catalog.DocumentID, docName constraints.Label) weblinking.Link {
	return weblinking.New(weblinking.Up, TextsItemPath(docID)).WithTitle(string(docName))
}

func itemLinkAbout(docID catalog.DocumentID, docName constraints.Label) weblinking.Link {
	return weblinking.New(weblinking.About, TextsItemPath(docID)).WithTitle(string(docName))
}

func (h *handler) registerItem(mux *http.ServeMux) {
	mux.Handle("GET /texts/{textID}",
		transport.NoBody(
			h.item,
		))
}

func (h *handler) initItem(base *ui.BaseTemplates) (err error) {
	h.itemTemplate, err = base.BuildHTML(deleteFormPartialTemplate, templateStatusPartialsHTML, templateItem)
	return
}

func (h *handler) item(w http.ResponseWriter, r *http.Request) {

	responseType, proceed := transport.NegotiateOr406(w, r, collectionLink(i18n.LangTagFromContext(r.Context())), supportedTypesItem)
	if !proceed {
		return
	}

	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}

	etagStrong := itemETagStrong(doc)
	setItemHeaders(w.Header())

	if responseType == mediatypes.JSON && transport.EtagMatchStrong(r.Header.Values("If-None-Match"), etagStrong) {
		respondItemNotModified(r.Context(), w, docID)
		return
	}

	langTag := i18n.LangTagFromContext(r.Context())

	var err error
	var output []byte
	switch responseType {
	case mediatypes.HTML:
		var errPrecondition *transport.ErrorPrecondition
		if noPre := consumeFlash(w, r, TextsItemPath(docID), flashDeleteNoPrecondition); noPre != "" {
			errPrecondition = new(transport.ErrPrecondition(http.StatusPreconditionRequired, langTag))
		} else if preFailed := consumeFlash(w, r, TextsItemPath(docID), flashDeletePreconditionFailed); preFailed != "" {
			errPrecondition = new(transport.ErrPrecondition(http.StatusPreconditionFailed, langTag))
		}
		if errPrecondition != nil {
			h.respondItemPreconditionHTML(w, r, docID, doc, errPrecondition, etagStrong)
			return
		}
		output, err = h.bufferItemHTML(w, r, doc, docID, etagStrong, i18n.LangTagFromContext(r.Context()), errPrecondition)
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
	h.Set("Allow", "GET, HEAD, PATCH, POST, DELETE, OPTIONS")
	h.Set("Accept-Patch", mediatypes.MergePatchJSON.String())
	h.Set("Accept-Post", mediatypes.MultipartForm.String())
}

func (h *handler) respondItemPreconditionHTML(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, doc catalog.Document, errPrecondition *transport.ErrorPrecondition, etagStrong transport.ETagStrong) {

	var output []byte
	var err error

	langTag := i18n.LangTagFromContext(r.Context())

	output, err = h.bufferItemHTML(w, r, doc, docID, etagStrong, langTag, errPrecondition)
	if err != nil {
		respondInternalError(w, r, err)
		return
	}

	transport.SetErrorHeaders(w.Header())

	writeItem(r.Context(), w, docID, doc, output, http.StatusConflict, langTag, mediatypes.HTML)
}

func respondItemNotModified(ctx context.Context, w http.ResponseWriter, docID catalog.DocumentID) {
	w.Header().Set("Link",
		append(
			weblinking.AllAlternates(TextsItemPath(docID), supportedTypesItem...),
			itemLinks(docID)...,
		).String(),
	)
	transport.WriteNoBody(ctx, w, http.StatusNotModified)
}

func writeItem(ctx context.Context, w http.ResponseWriter, docID catalog.DocumentID, doc catalog.Document, output []byte, statuscode int, langTag language.Tag, responseType mediatypes.Type) {

	links := append(weblinking.Alternates(TextsItemPath(docID), langTag, responseType, supportedTypesItem...), itemLinks(docID)...)
	if responseType == mediatypes.HTML {
		w.Header().Set("ETag", string(transport.GetEtagWeak(output)))
	} else {
		w.Header().Set("ETag", string(itemETagStrong(doc)))
		links = append(links, itemLinkJSONDescribedBy()...)
	}
	if doc.Status == catalog.StatusComplete {
		links = append(
			links,
			weblinking.New(weblinking.Search, weblinking.RootPath.Join("suggestions")).
				WithQueryVal("textID", strconv.FormatInt(int64(docID), 10)).
				WithTitle(i18n.Translate(langTag, i18n.SuggestionsTitleFilteredTextID, doc.Name)),
		)
	}
	w.Header().Set("Link", links.String())
	transport.Write(ctx, w, output, responseType, statuscode)
}

func (h *handler) bufferItemHTML(w http.ResponseWriter, r *http.Request, doc catalog.Document, docID catalog.DocumentID, etagStrong transport.ETagStrong, langTag language.Tag, errPrecondition *transport.ErrorPrecondition) ([]byte, error) {
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
				weblinking.Breadcrumb(weblinking.RootPath, T(i18n.NavigationHome)),
				collectionLink(langTag),
			},
			Links: itemLinks(docID),
		},
		itemData:          newItemData(doc),
		Status:            newStatusData(doc.Status, getStatusTranslation(doc.Status, langTag)),
		ErrorPrecondition: errPrecondition,
		FileSizeFmt:       i18n.FormatBytesLocale(doc.Filesize, langTag),
		WordCountFmt:      getWordcountFmt(doc.WordCount, langTag),
		ETagStrong:        etagStrong,
		Flash:             consumeFlash(w, r, TextsItemPath(docID), flashCreated, flashUpdated),
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
	return weblinking.Links{
		weblinking.New(weblinking.Up, CollectionPath),
		weblinking.New(weblinking.Collection, CollectionPath),
		weblinking.New(weblinking.Self, TextsItemPath(docID)),
		weblinking.New(weblinking.Enclosure, TextsFilePath(docID)).
			WithType(mediatypes.TextPlain),
		weblinking.New(weblinking.EditMedia, TextsFilePath(docID)).
			WithType(mediatypes.TextPlain),
		weblinking.New(weblinking.Status, TextsStatusPath(docID)),
		weblinking.New(weblinking.EditForm, TextsEditFormPath(docID)).
			WithType(mediatypes.HTML),
		weblinking.New(weblinking.DeleteForm, DeleteFormPath(docID)).
			WithType(mediatypes.HTML),
	}
}

func itemLinkJSONDescribedBy() weblinking.Links {
	return weblinking.Links{
		weblinking.New(weblinking.DescribedBy, ItemDocPath).WithType(mediatypes.HTML),
		weblinking.New(weblinking.DescribedBy, ItemSchemaPath).WithType(mediatypes.SchemaJSON),
		weblinking.New(weblinking.DescribedBy, ItemSpecPath).WithType(mediatypes.OpenAPIJSON),
	}
}

func respondMinimalItem(w http.ResponseWriter, ctx context.Context, doc catalog.Document, docID catalog.DocumentID, status int) {
	w.Header().Set("ETag", string(itemETagStrong(doc)))
	w.Header().Set("Link", append(itemLinks(docID), weblinking.AllAlternates(TextsItemPath(docID), supportedTypesItem...)...).String())
	transport.WriteNoBody(ctx, w, status)
}

// Verify at startup that itemData is valid for gob/ETag encoding
var _ = itemETagStrong(catalog.Document{})

func itemETagStrong(doc catalog.Document) transport.ETagStrong {
	etagStrong, err := transport.GetEtagStrong(newItemData(doc))
	if err != nil {
		panic(err)
	}
	return etagStrong
}

func getStatusTranslation(status catalog.SourceStatus, langTag language.Tag) string {
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
	return i18n.Translate(langTag, message)
}
