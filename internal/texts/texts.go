package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/constraints"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/weblinking"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	//advisory limit, recommend file upload instead if over
	textMaxWords     int   = 100000
	networkBufferMB  int64 = 5                                       // Other Fields + Protocol Overhead + small Buffer
	maxPayloadSizeMB int64 = catalog.MaxFileSizeMB + networkBufferMB // Network Limit
	maxPayloadSize   int64 = maxPayloadSizeMB * 1_000_000
)

var (
	postContentType mediatypes.Type = mediatypes.MultipartForm
)

type formFieldErrors map[string]string

func newFormFieldErrors(err error, source string, langTag language.Tag) (FieldErrors formFieldErrors) {

	FieldErrors = make(formFieldErrors)

	if err, ok := errors.AsType[*errs.InvalidParams](err); !ok {
		return
	} else {
		// if no known source selected => show only error to select a source, dont show errors of (non-visible until checked) text or file
		sourceSelected := slices.Contains(sources, source)

		for _, invalidParam := range *err {
			//if there are multiple, only put one error
			if _, ok := FieldErrors[invalidParam.Name]; !ok {
				if !sourceSelected && slices.Contains(sources, invalidParam.Name) {
					FieldErrors[fieldSourceName] = i18n.Func(langTag)(i18n.TextCreateFormUnselected)
					continue
				}
				FieldErrors[invalidParam.Name] = invalidParam.Reason
			}
		}
	}
	return
}

type FormUserData struct {
	Name   string
	Source string
	Text   string
}

type formData struct {
	FormUserData
	//Errors
	Errors formFieldErrors
	//constraints
	NameMaxCodePoints int
	WordMax           int
	WordMaxFmt        string
	WordMin           int
	FilesizeLimit     int64
	FilesizeLimitMB   int64
}

func newFormData(formUserData FormUserData, err error, lang language.Tag) formData {
	return formData{
		FormUserData:      formUserData,
		Errors:            newFormFieldErrors(err, formUserData.Source, lang),
		NameMaxCodePoints: catalog.DocumentNameMaxCodePoints,
		WordMax:           textMaxWords,
		WordMaxFmt:        message.NewPrinter(lang).Sprintf("%d", textMaxWords),
		WordMin:           catalog.MinWordsContents,
		FilesizeLimit:     catalog.MaxFileSize,
		FilesizeLimitMB:   catalog.MaxFileSizeMB,
	}
}

func respondNotFound(w http.ResponseWriter, r *http.Request) {
	transport.NotFound(w, r, collectionLink(i18n.LangTagFromContext(r.Context())))
}

func respondInternalError(w http.ResponseWriter, r *http.Request, err error) {
	transport.InternalError(w, r, err, collectionLink(i18n.LangTagFromContext(r.Context())))
}

func respondInternalErrorSubresource(w http.ResponseWriter, r *http.Request, err error, docID catalog.DocumentID, docName catalog.DocumentName) {
	transport.InternalError(w, r, err, itemLinkUp(docID, docName))
}

func respondBodyNotParseableItem(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, docName catalog.DocumentName) {
	transport.BadRequest(w, r, i18n.TranslateFuncFromContext(r.Context())(i18n.HTTPErrorBodyNotParseable), itemLinkUp(docID, docName))
}

func itemETagStrong(doc catalog.Document) (transport.ETagStrong, error) {
	return transport.GetEtagStrong(newItemData(doc))
}

func itemEtagStrongOr500(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, doc catalog.Document) (transport.ETagStrong, bool) {
	if ETagStrong, err := itemETagStrong(doc); err != nil {
		w.Header().Del("Location")
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return "", false
	} else {
		return ETagStrong, true
	}
}

func collectionPath() *url.URL {
	return (&url.URL{Path: "/"}).JoinPath("texts")
}

func itemPath(docID catalog.DocumentID) *url.URL {
	return collectionPath().JoinPath(strconv.FormatInt(int64(docID), 10))
}

func collectionLink(langTag language.Tag) weblinking.Link {
	return weblinking.New(weblinking.Collection, collectionPath()).WithTitle(i18n.Func(langTag)(i18n.NavigationTextCollection))
}

func itemLinkUp(docID catalog.DocumentID, docName catalog.DocumentName) weblinking.Link {
	return weblinking.New(weblinking.Up, itemPath(docID)).WithTitle(string(docName))
}

const (
	//name of field naming the resource
	fieldNameName = "name"
	//name of the field for choosing to provide text(-area content) or file(-input upload)
	fieldSourceName = "source"
	//keep the text (edit-form)
	sourceKeep = "keep"
	//name of the field containing text (and value for "source" to choose text(-area content))
	fieldTextName = "text"
	//name of the field containing a file upload (and value for "source" to choose file(-input upload))
	fieldFileName = "file"
)

var sources = []string{
	sourceKeep,
	fieldTextName,
	fieldFileName,
}

func parseMultipartFormOrRespond(w http.ResponseWriter, r *http.Request, langTag language.Tag, escape weblinking.Link) (userData FormUserData, filename string, contents []byte, proceed bool) {
	T := i18n.Func(langTag)
	err := r.ParseMultipartForm(maxPayloadSize)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			transport.ContentTooLarge(w, r, i18n.FormatBytesLocale(maxPayloadSize, langTag), escape)
		} else {
			transport.BadRequest(w, r, T(i18n.HTTPErrorBodyNotParseable), escape)
		}
		return FormUserData{}, "", []byte{}, false
	}
	defer r.MultipartForm.RemoveAll()

	userData.Name = r.FormValue(fieldNameName)
	userData.Source = r.FormValue(fieldSourceName)
	userData.Text = r.FormValue(fieldTextName)

	switch userData.Source {
	case fieldTextName:
		contents = []byte(userData.Text)
	case "", fieldFileName:
		file, header, err := r.FormFile(fieldFileName)
		if err != nil && !errors.Is(err, http.ErrMissingFile) {
			transport.InternalError(w, r, err, escape)
			return FormUserData{}, "", []byte{}, false
		} else if err == nil {
			defer file.Close()
			limitedFile := io.LimitReader(file, catalog.MaxFileSize+1)
			contents, err = io.ReadAll(limitedFile)
			if err != nil {
				transport.InternalError(w, r, err, escape)
				return FormUserData{}, "", []byte{}, false
			}
			filename = header.Filename
		} else {
			contents = []byte(userData.Text)
		}
	}
	return userData, strings.TrimSuffix(filename, ".txt"), contents, true
}

func (h *handler) getDocOrRespond(w http.ResponseWriter, r *http.Request, raw string) (docID catalog.DocumentID, doc catalog.Document, proceeed bool) {
	if docID, proceeed = catalog.ParseDocumentIDFromString(raw); !proceeed {
		respondNotFound(w, r)
		return
	}

	var err error
	if doc, err = h.service.GetDocument(docID); err != nil {
		var errWithCode *errs.CodedError
		if errors.As(err, &errWithCode) && errWithCode.Code() == errs.CodeNotFound {
			respondNotFound(w, r)
		} else {
			respondInternalError(w, r, err)
		}
	}
	proceeed = err == nil
	return
}

func mapCatalogErrors(langTag language.Tag, FieldErrors *errs.FieldErrors) *errs.InvalidParams {
	var iParams errs.InvalidParams

	for _, fieldErr := range FieldErrors.Errors() {
		iParams = append(iParams, mapCatalogError(langTag, fieldErr))
	}

	return &iParams
}

func mapCatalogError(langTag language.Tag, FieldError *errs.FieldError) *errs.InvalidParam {
	switch FieldError.Name {
	case "name":
		return mapNameError(langTag, FieldError)
	case "filename":
		return mapFilenameError(langTag, FieldError)
	case "file":
		return mapFileError(langTag, FieldError)
	case "text":
		return mapTextError(langTag, FieldError)
	}
	return nil
}

func mapFilenameError(langTag language.Tag, fieldErr *errs.FieldError) *errs.InvalidParam {
	T := i18n.Func(langTag)
	out := &errs.InvalidParam{
		Name:   fieldErr.Name,
		Code:   fieldErr.Code(),
		Params: fieldErr.Params(),
	}
	switch fieldErr.Code() {
	case errs.CodeMinCodePoints:
		out.Reason = T(i18n.TextFieldFilenameMinCodePoints, []any{out.Params[errs.ParamLimit]})

	case errs.CodeMaxBytes:
		out.Reason = T(i18n.TextFieldFilenameMaxBytes, []any{out.Params[errs.ParamLimit]})

	case errs.CodePattern:
		out.Reason = T(i18n.TextFieldFilenamePattern)

	case errs.CodeInvalid:
		p := fieldErr.Params()
		if p[errs.ParamEncoding] == "utf-8" {
			out.Reason = T(i18n.TextFieldFileInvalidUTF8)
		} else if p[errs.ParamRule] == errs.RuleReserved {
			out.Reason = T(i18n.TextFieldFilenameReserved)
		} else {
			out.Reason = T(i18n.CommonErrorUnknown)
		}

	default:
		out.Reason = T(i18n.CommonErrorUnknown)
	}
	return out
}

func mapNameError(langTag language.Tag, fieldErr *errs.FieldError) *errs.InvalidParam {
	T := i18n.Func(langTag)
	out := errs.InvalidParam{
		Name:   fieldErr.Name,
		Code:   fieldErr.Code(),
		Params: fieldErr.Params(),
	}
	switch fieldErr.Code() {
	case errs.CodeMaxCodePoints:
		out.Reason = T(i18n.TextFieldNameMaxCodePoints)
	case errs.CodeRequired:
		out.Reason = T(i18n.TextFieldNameRequired)
	}
	return &out
}

func mapTextError(langTag language.Tag, fieldErr *errs.FieldError) *errs.InvalidParam {
	T := i18n.Func(langTag)
	var reason string
	switch fieldErr.Code() {
	case errs.CodeRequired:
		reason = T(i18n.TextFieldTextMissing)
	case errs.CodeMinWords:
		reason = T(i18n.TextFieldTextMinWords, fieldErr.Params()[errs.ParamLimit])
	case errs.CodeMaxBytes:
		var limitMB int64
		if val, ok := fieldErr.Params()[errs.ParamLimit].(int64); ok {
			limitMB = val / 1_000_000
		} else {
			limitMB = catalog.MaxFileSizeMB
		}
		reason = T(i18n.TextFieldTextMaxBytes, limitMB)
	case errs.CodeInvalid:
		reason = T(i18n.TextFieldTextInvalidUTF8)
	}
	return errs.InvalidParamFromCodedError("text", fieldErr.CodedError, reason)
}

func mapFileError(langTag language.Tag, fieldErr *errs.FieldError) *errs.InvalidParam {
	T := i18n.Func(langTag)
	reason := T(i18n.CommonErrorUnknown)
	switch fieldErr.Code() {
	case errs.CodeRequired:
		reason = T(i18n.TextFieldFileMissing)
	case errs.CodeMinWords:
		reason = T(i18n.TextFieldFileMinWords, fieldErr.Params()[errs.ParamLimit])
	case errs.CodeMaxBytes:
		var limitMB int64
		if val, ok := fieldErr.Params()[errs.ParamLimit].(int64); ok {
			limitMB = val / 1_000_000
		} else {
			limitMB = catalog.MaxFileSizeMB
		}
		reason = T(i18n.TextFieldFileMaxBytes, limitMB)
	case errs.CodeInvalid:
		reason = T(i18n.TextFieldFileInvalidUTF8)
	}
	return errs.InvalidParamFromCodedError("file", fieldErr.CodedError, reason)
}

func deriveFilename(docName catalog.DocumentName, fileheaderName string, contents []byte) constraints.SafeFilename {
	return constraints.DeriveSafeFilename(string(docName), strings.TrimSuffix(fileheaderName, ".txt"), constraints.FirstBytes(string(contents[:min(50, len(contents))]), 50))
}
