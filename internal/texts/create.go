package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/weblinking"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
)

const (
	flashValueCreated string = "text.created"
)

func (h *handler) registerCreate(mux *http.ServeMux) {
	mux.Handle("POST /texts",
		transport.LimitBody(
			transport.ParsePreference(h.create),
			mediatypes.Types{mediatypes.MultipartForm}, maxPayloadSize),
	)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)
	T := i18n.Func(langTag)
	collection := weblinking.New(weblinking.Collection, collectionPath()).WithTitle(T(i18n.NavigationTextCollection))

	formUserData, filenameRaw, contentsRaw, proceed := parseMultipartFormOrRespond(w, r, langTag, collection)
	if !proceed {
		return
	}

	var invalidParams errs.InvalidParams

	fail := func(formUserData FormUserData, invalidParams errs.InvalidParams) {
		h.createWithErrors(w, r, formUserData, &invalidParams)
	}

	if formUserData.Source != fieldFileName && formUserData.Source != fieldTextName && formUserData.Source != "" {
		fail(formUserData, errs.InvalidParams{errs.InvalidParamFromCodedError(
			fieldSourceName,
			errs.Enum([]any{fieldFileName, fieldTextName, ""}, formUserData.Source),
			T(i18n.TextCreateFormUnselected),
		)})
		return
	}

	var coded *errs.CodedError
	var docName catalog.DocumentName
	var srcContents catalog.SourceContents
	if docName, coded = catalog.ParseDocumentName(formUserData.Name); coded != nil {
		invalidParams = append(invalidParams, mapCatalogError(langTag, errs.FieldFromCoded("name", coded)))
	}
	if srcContents, coded = catalog.ParseSourceContents(contentsRaw); coded != nil {
		invalidParams = append(invalidParams, mapCatalogError(langTag, errs.FieldFromCoded(formUserData.Source, coded)))
	}

	if len(invalidParams) > 0 {
		fail(formUserData, invalidParams)
	}

	doc, docID, err := h.service.CreateDocument(docName, deriveFilename(docName, filenameRaw, contentsRaw), srcContents)
	if err != nil {
		respondInternalError(w, r, err)
		return
	}
	ETagStrong, err := itemETagStrong(doc)
	if err != nil {
		slog.WarnContext(r.Context(), fmt.Sprintf("cannot create ETagStrong from document(%v): %v", doc, err))
	}

	if transport.IsFullDocumentNavigation(r) {
		http.SetCookie(w, &http.Cookie{
			Name:   string(flashCreated),
			Value:  flashValueCreated,
			MaxAge: 60,
			Path:   itemPath(docID).String(),
		})
		transport.SeeOther(w, r, itemPath(docID), T(i18n.TextCreateFlashSuccess))
		return
	}

	w.Header().Set("Location", itemPath(docID).String())

	negotiated := transport.NegotiateOrFallback(r, supportedTypesItem, mediatypes.HTML)

	preferences, ok := transport.GetPreferences(r.Context())
	if ok && slices.Contains(preferences, transport.ReturnMinimal) {
		transport.ApplyPreferences(w.Header(), transport.ReturnMinimal)
		w.Header().Set("Link", append(itemLinks(docID), itemLinkJSONDescribedBy()...).String())
		w.Header().Del("Content-Type")
		w.Header().Del("Content-Language")
		w.WriteHeader(http.StatusCreated)
		return
	}

	if preferences != nil && slices.Contains(preferences, transport.ReturnRepresentation) {
		transport.ApplyPreferences(w.Header(), transport.ReturnRepresentation)
	}

	var output []byte
	var ETagStr string

	switch negotiated {
	case mediatypes.HTML:
		output, err = h.bufferItemHTML(w, r, doc, docID, ETagStrong, langTag, nil)
		ETagStr = string(transport.GetEtagWeak(output))
	case mediatypes.JSON:
		output, err = bufferItemJSON(doc)
		ETagStr = string(ETagStrong)
	}
	if err != nil {
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return
	}

	w.Header().Set("Etag", ETagStr)

	writeItem(r.Context(), w, docID, doc, output, http.StatusCreated, langTag, negotiated)
}

func (h *handler) createWithErrors(w http.ResponseWriter, r *http.Request, formUserData FormUserData, invalidParams *errs.InvalidParams) {

	transport.SetErrorHeaders(w.Header())

	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)
	T := i18n.Func(langTag)

	responseType := transport.ResolveErrorContentType(r)

	var output []byte
	var err error
	switch responseType {
	case mediatypes.HTML:
		output, err = h.bufferCreateForm(ctx, newFormData(formUserData, invalidParams, langTag))
	case mediatypes.JSON:
		fallthrough
	case mediatypes.ProblemJSON:
		output, err = transport.ProblemValidation(r.Context(), invalidParams, "")
		w.Header().Set("Link", transport.GetProblemDescribedBy(transport.ProblemTypeValidation).String())
	case mediatypes.TextPlain:
		output = fmt.Appendf(nil, "%s\n%s", T(i18n.CommonErrorSummaryTitle), invalidParams.String())
	}

	if err != nil {
		transport.InternalError(w, r, err, weblinking.New(weblinking.Collection, collectionPath()).WithTitle(i18n.Func(langTag)(i18n.NavigationTextCollection)))
		return
	}

	w.Header().Set("Content-Type", responseType.String())
	w.WriteHeader(http.StatusUnprocessableEntity)

	transport.Write(ctx, w, output)
}
