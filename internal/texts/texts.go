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

func newFormFieldErrors(err error, src source, langTag language.Tag) (FieldErrors formFieldErrors) {

	FieldErrors = make(formFieldErrors)

	if err, ok := errors.AsType[*errs.FieldErrors](err); !ok {
		return
	} else {
		// if no known source selected => show only error to select a source, dont show errors of (non-visible until checked) text or file
		sourceSelected := slices.Contains(validSources, src)

		for _, invalidParam := range *err {
			//if there are multiple, only put one error
			if _, ok := FieldErrors[invalidParam.Name]; !ok {
				if !sourceSelected && slices.Contains(validSources, source(invalidParam.Name)) {
					FieldErrors[string(sourceField)] = i18n.Func(langTag)(i18n.TextCreateFormUnselected)
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
	fileField = "file"
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

// explicit radio (unenhanced HTML) wins; otherwise infer
func resolveSource(formValue string) source {
	if slices.Contains(validSources, source(formValue)) {
		return source(formValue)
	}
	return ""
}

func parseMultipartFormOrRespond(w http.ResponseWriter, r *http.Request, langTag language.Tag, escape weblinking.Link, allowedSources sources) (userData FormUserData, filename string, contents []byte, errSource *errs.CodedError, proceed bool) {
	T := i18n.Func(langTag)
	err := r.ParseMultipartForm(maxPayloadSize)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			transport.ContentTooLarge(w, r, i18n.FormatBytesLocale(maxPayloadSize, langTag), escape)
		} else {
			transport.BadRequest(w, r, T(i18n.HTTPErrorBodyNotParseable), escape)
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
		multipartFile, header, err := r.FormFile(string(fileField))
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
			filename = strings.TrimSuffix(header.Filename, ".txt")

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
		var errWithCode *errs.CodedError
		if errors.As(err, &errWithCode) && errWithCode.Code() == errs.CodeNotFound {
			respondNotFound(w, r)
		} else {
			respondInternalError(w, r, err)
		}
	}
	proceed = err == nil
	return
}

func mapFilenameError(langTag language.Tag, errCoded errs.CodedError) errs.FieldError {
	T := i18n.Func(langTag)
	var reason string
	switch errCoded.Code() {
	case errs.CodeMinCodePoints:
		reason = T(i18n.TextFieldFilenameMinCodePoints, errCoded.Params()[errs.ParamLimit])
	case errs.CodeMaxBytes:
		reason = T(i18n.TextFieldFilenameMaxBytes, errCoded.Params()[errs.ParamLimit])
	case errs.CodePattern:
		reason = T(i18n.TextFieldFilenamePattern)
	case errs.CodeInvalid:
		p := errCoded.Params()
		if p[errs.ParamEncoding] == "utf-8" {
			reason = T(i18n.TextFieldFileInvalidUTF8)
		} else {
			reason = T(i18n.CommonErrorUnknown)
		}
	default:
		reason = T(i18n.CommonErrorUnknown)
	}
	return errs.InvalidParamFromCodedError("filename", errCoded, reason)
}

func mapNameError(langTag language.Tag, errCoded errs.CodedError) errs.FieldError {
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
	return errs.InvalidParamFromCodedError(string(nameField), errCoded, reason)
}

func mapTextError(langTag language.Tag, errCoded errs.CodedError) errs.FieldError {
	T := i18n.Func(langTag)
	var reason string
	switch errCoded.Code() {
	case errs.CodeRequired:
		reason = T(i18n.TextFieldTextMissing)
	case errs.CodeMinWords:
		reason = T(i18n.TextFieldTextMinWords, errCoded.Params()[errs.ParamLimit])
	case errs.CodeMaxBytes:
		var limitMB int64
		if val, ok := errCoded.Params()[errs.ParamLimit].(int64); ok {
			limitMB = val / 1_000_000
		} else {
			limitMB = constraints.MaxFileSizeMB
		}
		reason = T(i18n.TextFieldTextMaxBytes, limitMB)
	case errs.CodeInvalid:
		reason = T(i18n.TextFieldTextInvalidUTF8)
	default:
		reason = T(i18n.CommonErrorUnknown)
	}
	return errs.InvalidParamFromCodedError(textField, errCoded, reason)
}

func mapFileError(langTag language.Tag, errCoded errs.CodedError) errs.FieldError {
	T := i18n.Func(langTag)
	reason := T(i18n.CommonErrorUnknown)
	switch errCoded.Code() {
	case errs.CodeRequired:
		reason = T(i18n.TextFieldFileMissing)
	case errs.CodeMinWords:
		reason = T(i18n.TextFieldFileMinWords, errCoded.Params()[errs.ParamLimit])
	case errs.CodeMaxBytes:
		var limitMB int64
		if val, ok := errCoded.Params()[errs.ParamLimit].(int64); ok {
			limitMB = val / 1_000_000
		} else {
			limitMB = constraints.MaxFileSizeMB
		}
		reason = T(i18n.TextFieldFileMaxBytes, limitMB)
	case errs.CodeInvalid:
		reason = T(i18n.TextFieldFileInvalidUTF8)
	}
	return errs.InvalidParamFromCodedError(fileField, errCoded, reason)
}

func deriveFilename(docName constraints.Label, fileheaderName string, contents []byte) constraints.SafeFilename {
	return constraints.DeriveSafeFilename(
		string(docName),
		strings.TrimSuffix(fileheaderName, ".txt"),
		constraints.FirstBytes(string(contents[:min(50, len(contents))]), 50),
	)
}
