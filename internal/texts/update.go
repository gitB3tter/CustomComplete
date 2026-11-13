package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/constraints"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/schema"
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
	mux.Handle("POST /texts/{textID}",
		transport.LimitBody(
			h.update,
			mediatypes.Types{mediatypes.MultipartForm}, maxPayloadSize),
	)
	mux.Handle("PATCH /texts/{textID}",
		transport.LimitBody(
			h.patch,
			mediatypes.Types{mediatypes.MergePatchJSON}, transport.MB(1)),
	)
	mux.Handle("PUT /texts/{textID}/file",
		transport.LimitBody(
			h.replaceFile,
			mediatypes.Types{mediatypes.TextPlain}, constraints.MaxFileSize),
	)
}

func (h *handler) update(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)

	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}
	itemLink := weblinking.New(weblinking.Up, TextsItemPath(docID)).WithTitle(string(doc.Name))

	userData, filenameRaw, contents, errSource, proceed := parseMultipartFormOrRespond(w, r, langTag, itemLink, validSources)
	if !proceed {
		return
	}

	T := i18n.Func(langTag)

	var errCoded *errs.CodedError
	var errsField errs.FieldErrors
	var srcContents catalog.SourceContents
	if errSource != nil {
		errsField.AddCoded(sourceField, *errSource, T(i18n.TextCreateFormUnselected))
	} else if userData.Source == keep {
	} else if srcContents, errCoded = catalog.ParseSourceContents(contents); errCoded != nil {
		switch userData.Source {
		case file:
			errsField = append(errsField, mapFileError(langTag, *errCoded))
		case text:
			errsField = append(errsField, mapTextError(langTag, *errCoded))
		default:
			errsField.AddCoded(sourceField, errs.Enum(validSources.toAny(), userData.Source), T(i18n.TextCreateFormUnselected))
		}
	}

	etagStrong := itemETagStrong(doc)

	respondError := func(err error) {
		h.respondUpdateErrors(w, r, docID, userData, doc, err, etagStrong)
	}

	if errPrecondition := transport.Precondition(etagStrong, r.FormValue(ui.HiddenFieldEtag), langTag); errPrecondition != nil {
		respondError(errPrecondition)
		return
	}
	patchMeta := catalog.DocumentPatchMetadata{}

	var docName constraints.Label
	if docName, errCoded = constraints.ParseLabel(userData.Name); errCoded != nil {
		errsField = append(errsField, mapNameError(langTag, *errCoded))
	} else {
		patchMeta.Name = &docName
	}

	if len(errsField) > 0 {
		respondError(&errsField)
		return
	}

	var err error
	if userData.Source != keep {
		patchMeta.Filename = new(deriveFilename(docName, filenameRaw, contents))

		doc, err = h.service.UpdateDocument(docID, patchMeta, srcContents)
	} else {
		doc, err = h.service.PatchDocumentMetadata(docID, patchMeta)
	}
	if err != nil {
		respondError(err)
		return
	}

	h.respondUpdateSuccess(w, r, docID, doc, false)
}

func (h *handler) patch(w http.ResponseWriter, r *http.Request) {

	var proceed bool

	docID, doc, proceed := h.getDocOrRespond(w, r)
	if !proceed {
		return
	}

	etagStrong := itemETagStrong(doc)

	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)
	T := i18n.Func(langTag)

	if errPrecondition := transport.Precondition(etagStrong, r.Header.Get("If-Match"), langTag); errPrecondition != nil {
		h.respondUpdateErrors(w, r, docID, FormUserData{}, doc, errPrecondition, etagStrong)
		return
	}

	var mergePatch map[string]json.RawMessage
	body, err := io.ReadAll(r.Body)
	if err == nil {
		err = json.Unmarshal(body, &mergePatch)
	}
	if err != nil {
		respondBodyNotParseableItem(w, r, docID, doc.Name)
		return
	}

	handleLenient := slices.Contains(transport.ParsePreferences(r.Header), transport.HandlingLenient)

	var invalidParams errs.FieldErrors
	//writeable field -> type-error message
	write := map[string]string{"name": T(i18n.TextFieldNameType), "filename": T(i18n.TextFieldFilenameType)}
	writeRaw := make(map[string]string, len(write))
	//readonly field -> error message
	readOnly := map[string]string{"filesize": T(i18n.TextFieldFilesizeReadOnly), "filehash": T(i18n.TextFieldFilehashReadOnly)}
	for key, raw := range mergePatch {
		message, ok := write[key]
		if ok {
			switch actual := schema.JsonType(raw); actual {
			case schema.String:
				var str string
				if err := json.Unmarshal(raw, &str); err != nil {
					respondBodyNotParseableItem(w, r, docID, doc.Name)
					return
				}
				writeRaw[key] = str
			case schema.Null:
				if key == "name" {
					message = T(i18n.TextFieldNameRequired)
				} else {
					message = T(i18n.TextFieldFileNameRequired)
				}
				invalidParams.AddCoded(key, errs.Required(), message)
			default:
				invalidParams.AddCoded(key, errs.Type(schema.String, actual), message)
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
			invalidParams = append(invalidParams, errs.InvalidParamFromCodedError(key, errs.AdditionalProperties(), T(i18n.CommonErrorAdditionalProperties)))
		}
	}

	patch := catalog.DocumentPatchMetadata{}

	if rawFilename, ok := writeRaw["filename"]; ok {
		var filename constraints.SafeFilename
		if handleLenient {
			filename, err = constraints.CanonicalSafeFilename(rawFilename)
		} else {
			filename, err = constraints.ParseSafeFilename(rawFilename)
		}
		if err != nil {
			if errCoded, ok := errors.AsType[*errs.CodedError](err); ok {
				invalidParams = append(invalidParams, mapFilenameError(langTag, *errCoded))
			} else {
				respondInternalErrorSubresource(w, r, err, docID, doc.Name)
				return
			}
		} else {
			patch.Filename = new(filename)
		}
	}

	rawName, ok := writeRaw["name"]
	if ok {
		docName, err := constraints.ParseLabel(rawName)
		if err != nil {
			invalidParams = append(invalidParams, mapNameError(langTag, *err))
		} else {
			patch.Name = new(docName)
		}
	}

	if len(invalidParams) > 0 {
		h.respondUpdateErrors(w, r, docID, FormUserData{Name: rawName}, doc, &invalidParams, etagStrong)
		return
	}

	newDoc := doc
	if !patch.Empty() {
		newDoc, err = h.service.PatchDocumentMetadata(docID, patch)
		if err != nil {
			userData := FormUserData{}
			if patch.Name != nil {
				userData.Name = string(*patch.Name)
			}
			h.respondUpdateErrors(w, r, docID, userData, doc, err, etagStrong)
			return
		}
	}

	h.respondUpdateSuccess(w, r, docID, newDoc, false)
}

func (h *handler) replaceFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)

	docID, doc, proceed := h.getDocOrRespond(w, r)
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
		h.respondUpdateErrors(w, r, docID, FormUserData{}, doc, &errs.FieldErrors{mapFileError(langTag, *errCoded)}, etagFile)
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
		weblinking.New(weblinking.Self, TextsItemPath(docID).Join("file")),
	}.String())
	transport.RespondNoContent(w)
}

func appendReadOnly[T comparable](handleLenient bool, params errs.FieldErrors, raw json.RawMessage, currentVal T, name string, message string) errs.FieldErrors {
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
	updatedItemPath := TextsItemPath(docID)
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

	preferences := transport.ParsePreferences(r.Header)
	if !connegOK || slices.Contains(preferences, transport.ReturnMinimal) {
		if connegOK && responseType == mediatypes.HTML {
			http.SetCookie(w, flashCookie)
		}
		if slices.Contains(preferences, transport.ReturnMinimal) {
			transport.ApplyPreference(w.Header(), transport.ReturnMinimal)
		}
		respondMinimalItem(w, ctx, doc, docID, http.StatusNoContent)
		return
	}

	if slices.Contains(preferences, transport.ReturnRepresentation) {
		transport.ApplyPreference(w.Header(), transport.ReturnRepresentation)
	}

	var output []byte
	var err error
	var ETagStr string
	newEtagStrong := itemETagStrong(doc)
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
		w.Header().Set("Content-Location", updatedItemPath.String())
	}
	writeItem(ctx, w, docID, doc, output, http.StatusOK, langTag, responseType)
}

func (h *handler) respondUpdateErrors(w http.ResponseWriter, r *http.Request, docID catalog.DocumentID, userData FormUserData, doc catalog.Document, err error, etagStrong transport.ETagStrong) {

	var output []byte

	var status int
	if errPre, ok := errors.AsType[*transport.ErrorPrecondition](err); ok {
		status = int(errPre.Status())
	} else if err, ok := errors.AsType[*errs.FieldErrors](err); ok {
		status = http.StatusUnprocessableEntity
	} else if errs.HasCode(err, errs.CodeNotFound) {
		respondNotFound(w, r)
		return
	} else {
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return
	}

	ctx := r.Context()
	responseType := transport.ResolveErrorContentType(r)

	langTag := i18n.LangTagFromContext(r.Context())
	var etagStr string
	switch responseType {
	case mediatypes.HTML:
		output, err = h.bufferEditForm(ctx, docID, newEditFormData(userData, doc, err, etagStrong, langTag))
		etagStr = string(transport.GetEtagWeak(output))
	case mediatypes.ProblemJSON, mediatypes.JSON:
		if errPre, ok := errors.AsType[*transport.ErrorPrecondition](err); ok {
			output, err = transport.ProblemGeneric(errPre.Error())
		} else if errInvalidParams, ok := errors.AsType[*errs.FieldErrors](err); ok {
			output, err = transport.ProblemFieldError(r.Context(), *errInvalidParams, "")
			w.Header().Set(
				"Link",
				transport.ProblemErrorsPath.
					Link(weblinking.DescribedBy).
					WithTitle(i18n.Translate(langTag, i18n.NavigationDocumentation)).String(),
			)
		}
		etagStr = string(etagStrong)
	case mediatypes.TextPlain:
		output = []byte(err.Error())
		etagStr = string(etagStrong)
	}
	if err != nil {
		slog.Warn("output response text item update", slog.Any("type", responseType.String()), slog.Any("error", err))
		output = []byte(err.Error())
		responseType = mediatypes.TextPlain
	}

	transport.SetErrorHeaders(w.Header())

	w.Header().Set("ETag", etagStr)

	transport.Write(ctx, w, output, responseType, status)
}
