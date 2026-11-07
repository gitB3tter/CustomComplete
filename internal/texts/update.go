package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/constraints"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"slices"
)

func (h *handler) registerUpdate(mux *http.ServeMux) {
	mux.Handle("POST /texts/{id}",
		transport.LimitBody(
			transport.ParsePreference(
				h.update,
			),
			mediatypes.Types{mediatypes.MultipartForm}, maxPayloadSize),
	)
	mux.Handle("PATCH /texts/{id}",
		transport.LimitBody(
			transport.ParsePreference(
				h.patch,
			),
			mediatypes.Types{mediatypes.MergePatchJSON}, transport.MB(1)),
	)
	mux.Handle("PUT /texts/{id}/file",
		transport.LimitBody(
			transport.ParsePreference(
				h.replaceFile,
			),
			mediatypes.Types{mediatypes.TextPlain}, catalog.MaxFileSize),
	)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)

	docID, doc, proceed := h.getDocOrRespond(w, r, r.PathValue("id"))
	if !proceed {
		return
	}
	itemLink := weblinking.New(weblinking.Up, itemPath(docID)).WithTitle(string(doc.Name))

	userData, filenameRaw, contentsRaw, proceed := parseMultipartFormOrRespond(w, r, langTag, itemLink)
	if !proceed {
		return
	}

	var ETagStrong transport.ETagStrong
	if ETagStrong, proceed = itemEtagStrongOr500(w, r, docID, doc); !proceed {
		return
	}

	respondError := func(err error) {
		h.respondUpdateErrors(w, r, docID, userData, doc, err, ETagStrong)
	}

	if errPreconditon := transport.Precondition(ETagStrong, r.FormValue(ui.HiddenFieldEtag), langTag); errPreconditon != nil {
		respondError(errPreconditon)
		return
	}

	fieldErrs := &errs.FieldErrors{}
	patchMeta := catalog.DocumentPatchMetadata{}

	var docName catalog.DocumentName
	var coded *errs.CodedError
	if docName, coded = catalog.ParseDocumentName(userData.Name); coded != nil {
		fieldErrs.Add("name", coded)
	} else {
		patchMeta.Name = &docName
	}

	allowedSources := append(sources, "")

	if !slices.Contains(allowedSources, userData.Source) {
		var sourcesAny []any
		for _, source := range allowedSources {
			sourcesAny = append(sourcesAny, source)
		}
		params := mapCatalogErrors(langTag, fieldErrs)
		*params = append(*params, errs.InvalidParamFromCodedError(
			fieldSourceName,
			errs.Enum(sourcesAny, userData.Source),
			i18n.Func(langTag)(i18n.TextCreateFormUnselected),
		))
		respondError(params)
		return
	}

	var err error
	if userData.Source != sourceKeep {
		patchMeta.Filename = new(deriveFilename(docName, filenameRaw, contentsRaw))

		var srcContents catalog.SourceContents
		if srcContents, coded = catalog.ParseSourceContents(contentsRaw); coded != nil {
			fieldErrs.Add(userData.Source, coded)
		}

		if !fieldErrs.OK() {
			respondError(mapCatalogErrors(langTag, fieldErrs))
			return
		}

		doc, err = h.service.UpdateDocument(docID, patchMeta, srcContents)
	} else {
		if !fieldErrs.OK() {
			respondError(mapCatalogErrors(langTag, fieldErrs))
			return
		}
		doc, err = h.service.PatchDocumentMetadata(docID, patchMeta)
	}
	if err != nil {
		if fieldErrs, ok := errors.AsType[*errs.FieldErrors](err); ok {
			respondError(mapCatalogErrors(langTag, fieldErrs))
		} else {
			respondError(err)
		}
		return
	}

	h.respondUpdateSuccess(w, r, docID, doc, false)
}

func (h *handler) patch(w http.ResponseWriter, r *http.Request) {

	var proceed bool

	docID, doc, proceed := h.getDocOrRespond(w, r, r.PathValue("id"))
	if !proceed {
		return
	}

	var ETagStrong transport.ETagStrong
	if ETagStrong, proceed = itemEtagStrongOr500(w, r, docID, doc); !proceed {
		return
	}

	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)
	T := i18n.Func(langTag)

	if errPreconditon := transport.Precondition(ETagStrong, r.Header.Get("If-Match"), langTag); errPreconditon != nil {
		h.respondUpdateErrors(w, r, docID, FormUserData{}, doc, errPreconditon, ETagStrong)
		return
	}

	var patch map[string]json.RawMessage
	body, err := io.ReadAll(r.Body)
	if err == nil {
		err = json.Unmarshal(body, &patch)
	}
	if err != nil {
		respondBodyNotParseableItem(w, r, docID, doc.Name)
		return
	}

	preferences, hasPreferences := transport.GetPreferences(ctx)
	handleLenient := hasPreferences && slices.Contains(preferences, transport.HandlingLenient)

	unmarshal := func(raw json.RawMessage) (any, bool) {
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			respondBodyNotParseableItem(w, r, docID, doc.Name)
			return v, false
		}
		return v, true
	}

	var invalidParams errs.InvalidParams
	//writeable field -> type-error message
	write := map[string]string{"name": T(i18n.TextFieldNameType), "filename": T(i18n.TextFieldFilenameType)}
	writeRaw := make(map[string]string, len(write))
	//readonly field -> error message
	readOnly := map[string]string{"filesize": T(i18n.TextFieldFilesizeReadOnly), "filehash": T(i18n.TextFieldFilehashReadOnly)}
	for key, raw := range patch {
		message, ok := write[key]
		if ok {
			val, ok := unmarshal(raw)
			if !ok {
				return
			}
			addTypeError := func(actual transport.JSONType) {
				invalidParams.AddCoded(key, transport.ErrJSONTyp(transport.JSONTypeString, actual), message)
			}
			switch t := val.(type) {
			case string:
				writeRaw[key] = t
			case float64:
				addTypeError(transport.JSONTypeNumber)
			case bool:
				addTypeError(transport.JSONTypeBoolean)
			case []any:
				addTypeError(transport.JSONTypeArray)
			case map[string]any:
				addTypeError(transport.JSONTypeObject)
			case nil:
				if key == "name" {
					message = T(i18n.TextFieldNameRequired)
				} else {
					message = T(i18n.TextFieldFileNameRequired)
				}
				invalidParams.AddCoded(key, errs.Required(), message)
			default:
				addTypeError(transport.JSONTypeInvalid)
			}
			continue
		}
		message, ok = readOnly[key]
		if ok {
			switch key {
			case "filesize":
				invalidParams = appendReadOnly(handleLenient, invalidParams, raw, doc.Filesize, key, message)
				continue
			case "filehash":
				invalidParams = appendReadOnly(handleLenient, invalidParams, raw, doc.Hash, key, message)
				continue
			}
			continue
		}
		if !handleLenient {
			invalidParams = append(invalidParams, errs.InvalidParamFromCodedError(key, errs.Unexpected(), T(i18n.CommonErrorAdditionalProperties)))
		}
	}

	update := catalog.DocumentPatchMetadata{}

	if rawFilename, ok := writeRaw["filename"]; ok {
		var filename constraints.SafeFilename
		if handleLenient {
			filename, err = constraints.CanonicalSafeFilename(rawFilename)
		} else {
			filename, err = constraints.ParseSafeFilename(rawFilename)
		}
		if err != nil {
			if coded, ok := errors.AsType[*errs.CodedError](err); ok {
				invalidParams = append(invalidParams, mapFilenameError(langTag, errs.FieldFromCoded("filename", coded)))
			} else {
				respondInternalErrorSubresource(w, r, err, docID, doc.Name)
				return
			}
		} else {
			update.Filename = new(filename)
		}
	}

	rawName, ok := writeRaw["name"]
	if ok {
		docName, err := catalog.ParseDocumentName(rawName)
		if err != nil {
			if coded, ok := errors.AsType[*errs.CodedError](err); ok {
				invalidParams = append(invalidParams, mapNameError(langTag, errs.FieldFromCoded("name", coded)))
			} else {
				respondInternalErrorSubresource(w, r, err, docID, doc.Name)
				return
			}
		} else {
			update.Name = new(docName)
		}
	}

	if len(invalidParams) > 0 {
		h.respondUpdateErrors(w, r, docID, FormUserData{Name: rawName}, doc, &invalidParams, ETagStrong)
		return
	}

	docNew := doc
	if !update.Empty() {
		docNew, err = h.service.PatchDocumentMetadata(docID, update)
		if err != nil {
			userData := FormUserData{}
			if update.Name != nil {
				userData.Name = string(*update.Name)
			}
			if fieldErrs, ok := errors.AsType[*errs.FieldErrors](err); ok {
				err = mapCatalogErrors(langTag, fieldErrs)
			}
			h.respondUpdateErrors(w, r, docID, userData, doc, err, ETagStrong)
			return
		}
	}

	h.respondUpdateSuccess(w, r, docID, docNew, false)
}

func (h *handler) replaceFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)

	docID, doc, proceed := h.getDocOrRespond(w, r, r.PathValue("id"))
	if !proceed {
		return
	}

	etagFile := transport.GetEtagStrongLiteral(string(doc.Hash))

	if errPrecondition := transport.Precondition(etagFile, r.Header.Get("If-Match"), langTag); errPrecondition != nil {
		h.respondUpdateErrors(w, r, docID, FormUserData{}, doc, errPrecondition, etagFile)
		return
	}

	T := i18n.Func(langTag)
	contents, err := io.ReadAll(r.Body)
	if err != nil {
		transport.BadRequest(w, r, T(i18n.HTTPErrorBodyNotParseable), itemLinkUp(docID, doc.Name))
		return
	}
	srcContents, errCoded := catalog.ParseSourceContents(contents)
	if errCoded != nil {
		h.respondUpdateErrors(w, r, docID, FormUserData{}, doc, errCoded, etagFile)
		return
	}

	newDoc, err := h.service.UpdateDocument(docID, catalog.DocumentPatchMetadata{}, srcContents)
	if err != nil {
		h.respondUpdateErrors(w, r, docID, FormUserData{}, doc, err, etagFile)
		return
	}

	newETagFile := transport.GetEtagStrongLiteral(string(newDoc.Hash))
	w.Header().Set("ETag", string(newETagFile))
	w.Header().Set("Link", weblinking.Links{
		itemLinkUp(docID, newDoc.Name),
		weblinking.New(weblinking.Self, itemPath(docID).JoinPath("file")),
	}.String())
	transport.RespondNoContent(w)
}

func appendReadOnly[T comparable](handleLenient bool, params errs.InvalidParams, raw json.RawMessage, currentVal T, name string, message string) errs.InvalidParams {
	if handleLenient {
		var clientVal T
		if err := json.Unmarshal(raw, &clientVal); err != nil || clientVal != currentVal {
			params.AddCoded(name, errs.ReadOnly(), message)
		}
	} else {
		params.AddCoded(name, errs.ReadOnly(), message)
	}
	return params
}

func (h *handler) respondUpdateSuccess(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, doc catalog.Document, setContentLocation bool) {
	updatedItemPath := itemPath(docID)
	ctx := r.Context()

	flashCookie := &http.Cookie{
		Name:   string(flashUpdated),
		MaxAge: 60,
		Path:   updatedItemPath.String(),
	}

	langTag := i18n.LangTagFromContext(ctx)
	T := i18n.Func(langTag)

	if transport.IsFullDocumentNavigation(r) {
		http.SetCookie(w, flashCookie)
		transport.SeeOther(w, r, updatedItemPath, T(i18n.TextUpdateFlashSuccess))
		return
	}

	requested := mediatypes.ParseAccept(r.Header.Values("Accept"))
	responseType, connegOK := supportedTypesItem.Negotiate(requested)

	setItemHeaders(w.Header())

	preferences, hasPreferences := transport.GetPreferences(ctx)
	if !connegOK || hasPreferences && slices.Contains(preferences, transport.ReturnMinimal) {
		if connegOK && responseType == mediatypes.HTML {
			http.SetCookie(w, flashCookie)
		}
		if hasPreferences && slices.Contains(preferences, transport.ReturnMinimal) {
			transport.ApplyPreferences(w.Header(), transport.ReturnMinimal)
		}
		respondMinimalItem(w, ctx, doc, docID, http.StatusNoContent)
		return
	}

	if hasPreferences && slices.Contains(preferences, transport.ReturnRepresentation) {
		transport.ApplyPreferences(w.Header(), transport.ReturnRepresentation)
	}

	var output []byte
	var ETagStr string

	newEtagStrong, err := itemETagStrong(doc)
	if err != nil {
		slog.Warn("generate new ETagStrong", slog.Any("err", err))
		newEtagStrong = ""
	}

	switch responseType {
	case mediatypes.HTML:
		output, err = h.bufferItemHTML(w, r, doc, docID, newEtagStrong, langTag, nil)
		ETagStr = string(transport.GetEtagWeak(output))
	case mediatypes.JSON:
		output, err = bufferItemJSON(doc)
		ETagStr = string(newEtagStrong)
	}
	if err != nil {
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return
	}

	if ETagStr != "" {
		w.Header().Set("ETag", ETagStr)
	}

	if setContentLocation {
		w.Header().Set("Content-Location", updatedItemPath.Path)
	}
	writeItem(ctx, w, docID, doc, output, http.StatusOK, langTag, responseType)
}

func (h *handler) respondUpdateErrors(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, userData FormUserData, doc catalog.Document, err error, ETagStrong transport.ETagStrong) {

	var output []byte

	var status int
	if errPre, ok := errors.AsType[*transport.ErrorPrecondition](err); ok {
		status = int(errPre.Status())
	} else if err, ok := errors.AsType[*errs.InvalidParams](err); ok {
		status = http.StatusUnprocessableEntity
	} else {
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return
	}

	ctx := r.Context()
	responseType := transport.ResolveErrorContentType(r)

	var etagStr string
	switch responseType {
	case mediatypes.HTML:
		output, err = h.bufferEditForm(ctx, docID, newEditFormData(userData, doc, err, ETagStrong, i18n.LangTagFromContext(r.Context())))
		etagStr = string(transport.GetEtagWeak(output))
	case mediatypes.ProblemJSON, mediatypes.JSON:
		links := weblinking.Links{}
		if errPre, ok := errors.AsType[*transport.ErrorPrecondition](err); ok {
			output, err = transport.ProblemGeneric(errPre.Error())
		} else if errInvalidP, ok := errors.AsType[*errs.InvalidParams](err); ok {
			output, err = transport.ProblemValidation(r.Context(), errInvalidP, "")
			links = append(links, transport.GetProblemDescribedBy(transport.ProblemTypeValidation)...)
		}
		w.Header().Set("Link", links.String())
		etagStr = string(ETagStrong)
	case mediatypes.TextPlain:
		output = []byte(err.Error())
		etagStr = string(ETagStrong)
	}
	if err != nil {
		slog.Warn("buffer response text item update", slog.Any("type", responseType.String()), slog.Any("error", err))
		output = []byte(err.Error())
		responseType = mediatypes.TextPlain
	}

	transport.SetErrorHeaders(w.Header())

	w.Header().Set("ETag", etagStr)

	w.Header().Set("Content-Type", responseType.String())
	w.WriteHeader(status)
	transport.Write(ctx, w, output)
}
