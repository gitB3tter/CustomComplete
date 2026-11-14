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
	"slices"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	//advisory limit, recommend file upload instead if over
	textMaxWords     int   = 100000
	networkBufferMB  int64 = 5                                           // Other Fields + Protocol Overhead + small output
	maxPayloadSizeMB int64 = constraints.MaxFileSizeMB + networkBufferMB // Network Limit
	maxPayloadSize   int64 = maxPayloadSizeMB * 1_000_000
)

var (
	postContentType mediatypes.Type = mediatypes.MultipartForm
	specPath        weblinking.Path = weblinking.DocsSpecsPath.Join("texts")
	schemaPath      weblinking.Path = weblinking.DocsSchemasPath.Join("texts")
	docPath         weblinking.Path = weblinking.DocsPath.Join("texts")
)

type formFieldErrors map[string]string

func newFormFieldErrors(err error, src source, langTag language.Tag) (errsFormField formFieldErrors) {

	errsFormField = make(formFieldErrors)

	errsField, ok := errors.AsType[*errs.Fields](err)
	if !ok {
		return
	}
	// if no known source selected => show only error to select a source, dont show errors of (non-visible until checked) text or file
	sourceSelected := slices.Contains(validSources, src)

	for _, errField := range *errsField {
		//if there are multiple, only put one error
		if _, ok := errsFormField[errField.Name]; !ok {
			if !sourceSelected && slices.Contains(validSources, source(errField.Name)) {
				errsFormField[sourceField] = i18n.Translate(langTag, i18n.TextCreateFormUnselected)
				continue
			}
			errsFormField[errField.Name] = errField.Reason
		}
	}
	return
}

type FormUserData struct {
	Name   string
	Source source
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

func newFormData(userData FormUserData, err error, langTag language.Tag) formData {
	return formData{
		FormUserData:      userData,
		Errors:            newFormFieldErrors(err, userData.Source, langTag),
		NameMaxCodePoints: catalog.DocumentNameMaxCodePoints,
		WordMax:           textMaxWords,
		WordMaxFmt:        message.NewPrinter(langTag).Sprintf("%d", textMaxWords),
		WordMin:           catalog.MinWordsContents,
		FilesizeLimit:     constraints.MaxFileSize,
		FilesizeLimitMB:   constraints.MaxFileSizeMB,
	}
}

func respondNotFound(w http.ResponseWriter, r *http.Request) {
	transport.NotFound(w, r, collectionLink(i18n.LangTagFromContext(r.Context())))
}

func respondInternalError(w http.ResponseWriter, r *http.Request, err error) {
	transport.InternalError(w, r, err, collectionLink(i18n.LangTagFromContext(r.Context())))
}

func respondInternalErrorSubresource(w http.ResponseWriter, r *http.Request, err error, docID catalog.DocumentID, docName constraints.Label) {
	transport.InternalError(w, r, err, itemLinkUp(docID, docName))
}

func respondBodyNotParseableItem(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, docName constraints.Label) {
	transport.BadRequest(w, r, i18n.TranslateFuncFromContext(r.Context())(i18n.HTTPErrorBodyNotParseable), itemLinkUp(docID, docName))
}

type source string

type sources []source

func (srcs sources) toAny() (out []any) {
	out = make([]any, 0, len(srcs))
	for _, src := range srcs {
		out = append(out, src)
	}
	return
}

const (
	//nameField of field naming the resource
	nameField = "name"
	//name of the field for choosing to provide text(-area content) or file(-input upload)
	sourceField = "source"
	//name of field containing text (and value for "source" to choose text(-area content))
	textField = "text"
	//name of the field containing a file upload (and value for "source" to choose file(-input upload))
	fileField     = "file"
	filenameField = "filename"
	//keep the text (edit-form)
	keep source = "keep"
	text source = source(textField)
	file source = source(fileField)
)

var (
	validSources = sources{
		keep,
		text,
		file,
	}
)

func resolveSource(formValue string) source {
	if slices.Contains(validSources, source(formValue)) {
		return source(formValue)
	}
	return ""
}

func parseMultipartFormOrRespond(w http.ResponseWriter, r *http.Request, langTag language.Tag, escape weblinking.Link, allowedSources sources) (userData FormUserData, filename string, contents []byte, errSource error, proceed bool) {
	err := r.ParseMultipartForm(maxPayloadSize)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			transport.ContentTooLarge(w, r, i18n.FormatBytesLocale(maxPayloadSize, langTag), escape)
		} else {
			transport.BadRequest(w, r, i18n.Translate(langTag, i18n.HTTPErrorBodyNotParseable), escape)
		}
		proceed = false
		return
	}
	defer r.MultipartForm.RemoveAll()

	userData.Name = r.PostFormValue(nameField)
	userData.Source = resolveSource(r.PostFormValue(sourceField))
	userData.Text = r.PostFormValue(textField)

	switch userData.Source {
	case text:
		contents = []byte(userData.Text)
	case "", file:
		multipartFile, header, err := r.FormFile(fileField)
		if err != nil && !errors.Is(err, http.ErrMissingFile) {
			transport.InternalError(w, r, err, escape)
			return FormUserData{}, "", []byte{}, nil, false
		} else if err == nil {
			defer multipartFile.Close()
			if userData.Source == "" && len(userData.Text) > 0 {
				errSource = new(errs.Enum(allowedSources.toAny(), r.PostFormValue(sourceField)))
				proceed = true
				return
			}
			userData.Source = file
			limitedFile := io.LimitReader(multipartFile, constraints.MaxFileSize+1)
			contents, err = io.ReadAll(limitedFile)
			if err != nil {
				transport.InternalError(w, r, err, escape)
				return FormUserData{}, "", []byte{}, nil, false
			}
			filename = header.Filename

		} else if userData.Source == "" && len(userData.Text) > 0 {
			userData.Source = text
			contents = []byte(userData.Text)
		}
	}

	return userData, filename, contents, nil, true
}

func (h *handler) getDocOrRespond(w http.ResponseWriter, r *http.Request) (docID catalog.DocumentID, doc catalog.Document, proceed bool) {
	if docID, proceed = catalog.ParseDocumentIDFromString(r.PathValue("textID")); !proceed {
		respondNotFound(w, r)
		return
	}

	var err error
	if doc, err = h.service.GetDocument(docID); err != nil {
		if errs.HasCode(err, errs.CodeNotFound) {
			respondNotFound(w, r)
		} else {
			respondInternalError(w, r, err)
		}
	}
	proceed = err == nil
	return
}

func codedOrRespond(w http.ResponseWriter, r *http.Request, err error) (errs.Coded, bool) {
	errCoded, ok := errors.AsType[*errs.Coded](err)
	if !ok {
		respondInternalError(w, r, err)
	}
	return *errCoded, ok
}

func codedOrRespondSubresource(w http.ResponseWriter, r *http.Request, err error, docID catalog.DocumentID, docName constraints.Label) (errs.Coded, bool) {
	errCoded, ok := errors.AsType[*errs.Coded](err)
	if !ok {
		respondInternalErrorSubresource(w, r, err, docID, docName)
	}
	return *errCoded, ok
}

func mapFilenameError(langTag language.Tag, errCoded errs.Coded) errs.Field {
	T := i18n.Func(langTag)
	var reason string
	switch errCoded.Code() {
	case errs.CodeMinCodePoints:
		reason = T(i18n.TextFieldFilenameMinCodePoints, errs.ParamOr(errCoded, errs.Limit, constraints.SafeFileNameMinCodePoints))
	case errs.CodeMaxBytes:
		reason = T(i18n.TextFieldFilenameMaxBytes, errs.ParamOr(errCoded, errs.Limit, constraints.SafeFileNameMaxBytes))
	case errs.CodePattern:
		reason = T(i18n.TextFieldFilenamePattern)
	case errs.CodeInvalid:
		p := errCoded.Params()
		if p[errs.Encoding] == "utf-8" {
			reason = T(i18n.TextFieldFileInvalidUTF8)
		} else {
			reason = T(i18n.CommonErrorUnknown)
		}
	default:
		reason = T(i18n.CommonErrorUnknown)
	}
	return errs.FieldFromCoded(filenameField, errCoded, reason)
}

func mapNameError(langTag language.Tag, errCoded errs.Coded) errs.Field {
	T := i18n.Func(langTag)
	var reason string
	switch errCoded.Code() {
	case errs.CodeMaxCodePoints:
		reason = T(i18n.TextFieldNameMaxCodePoints)
	case errs.CodeRequired:
		reason = T(i18n.TextFieldNameRequired)
	default:
		reason = T(i18n.CommonErrorUnknown)
	}
	return errs.FieldFromCoded(nameField, errCoded, reason)
}

func mapSourceContentsError(langTag language.Tag, allowed sources, actual source, errCoded errs.Coded) errs.Field {
	switch actual {
	case file:
		return mapFileError(langTag, errCoded)
	case text:
		return mapContentsError(langTag, errCoded, i18n.TextFieldTextMissing, i18n.TextFieldTextMinWords, i18n.TextFieldTextMaxBytes, i18n.TextFieldTextInvalidUTF8)
	default:
		return errs.FieldFromCoded(sourceField, errs.Enum(allowed.toAny(), actual), i18n.Translate(langTag, i18n.TextCreateFormUnselected))
	}
}

func mapContentsError(langTag language.Tag, errCoded errs.Coded, required, minWords, maxBytes, invalid i18n.Key) errs.Field {
	T := i18n.Func(langTag)
	var reason string
	switch errCoded.Code() {
	case errs.CodeRequired:
		reason = T(required)
	case errs.CodeMinWords:
		reason = T(minWords, errs.ParamOr(errCoded, errs.Limit, catalog.MinWordsContents))
	case errs.CodeMaxBytes:
		reason = T(
			maxBytes,
			errs.ParamOr(errCoded, errs.Limit, constraints.MaxFileSize)/1_000_000,
		)
	case errs.CodeInvalid:
		if errCoded.Params()[errs.Encoding] == "utf-8" {
			reason = T(invalid)
		} else {
			reason = T(i18n.CommonErrorUnknown)
		}
	default:
		reason = T(i18n.CommonErrorUnknown)
	}
	return errs.FieldFromCoded(textField, errCoded, reason)
}

func mapFileError(langTag language.Tag, errCoded errs.Coded) errs.Field {
	return mapContentsError(langTag, errCoded, i18n.TextFieldFileMissing, i18n.TextFieldFileMinWords, i18n.TextFieldFileMaxBytes, i18n.TextFieldFileInvalidUTF8)
}

func deriveFilename(docName constraints.Label, fileheaderName string, contents []byte) constraints.SafeFilename {
	return constraints.DeriveSafeFilename(
		string(docName),
		strings.TrimSuffix(fileheaderName, ".txt"),
		constraints.FirstBytes(string(contents[:min(50, len(contents))]), 50),
	)
}
