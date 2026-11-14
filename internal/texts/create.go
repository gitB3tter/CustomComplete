package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/constraints"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/weblinking"
	"net/http"
	"slices"
)

const (
	flashValueCreated string = "text.created"
)

func (h *handler) registerCreate(mux *http.ServeMux) {
	mux.Handle("POST "+string(CollectionPath),
		transport.LimitBody(h.create,
			mediatypes.Types{mediatypes.MultipartForm},
			maxPayloadSize,
		),
	)
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)
	T := i18n.Func(langTag)

	var errsField errs.Fields
	allowedSources := sources{text, file}
	userData, filename, contents, errSource, proceed := parseMultipartFormOrRespond(w, r, langTag, collectionLink(langTag), allowedSources)
	if !proceed {
		return
	}
	var err error
	var srcContents catalog.SourceContents
	if errSource != nil {
		errCoded, ok := codedOrRespond(w, r, errSource)
		if !ok {
			return
		}
		errsField.AddCoded(sourceField, errCoded, T(i18n.TextCreateFormUnselected))
	} else {
		if srcContents, err = catalog.ParseSourceContents(contents); err != nil {
			errCoded, ok := codedOrRespond(w, r, err)
			if !ok {
				return
			}
			errsField = append(errsField, mapSourceContentsError(langTag, allowedSources, userData.Source, errCoded))
		}
	}

	var docName constraints.Label
	if docName, err = constraints.ParseLabel(userData.Name); err != nil {
		errCoded, ok := codedOrRespond(w, r, err)
		if !ok {
			return
		}
		errsField = append(errsField, mapNameError(langTag, errCoded))
	}

	if len(errsField) > 0 {
		h.respondCreateErrors(w, r, userData, errsField)
		return
	}

	doc, docID, err := h.service.CreateDocument(docName, deriveFilename(docName, filename, contents), srcContents)
	if err != nil {
		respondInternalError(w, r, err)
		return
	}

	if transport.IsFullDocumentNavigation(r) {
		http.SetCookie(w, &http.Cookie{
			Name:   string(flashCreated),
			Value:  flashValueCreated,
			MaxAge: 60,
			Path:   TextsItemPath(docID).String(),
		})
		transport.SeeOther(w, r, TextsItemPath(docID), T(i18n.TextCreateFlashSuccess))
		return
	}

	w.Header().Set("Location", TextsItemPath(docID).String())

	responseType := transport.NegotiateOrFallback(r, supportedTypesItem, mediatypes.HTML)

	preferences := transport.ParsePreferences(r.Header)
	if slices.Contains(preferences, transport.ReturnMinimal) {
		transport.ApplyPreference(w.Header(), transport.ReturnMinimal)
		w.Header().Set("Link", append(itemLinks(docID), itemLinkJSONDescribedBy()...).String())
		transport.WriteNoBody(r.Context(), w, http.StatusCreated)
		return
	}

	if slices.Contains(preferences, transport.ReturnRepresentation) {
		transport.ApplyPreference(w.Header(), transport.ReturnRepresentation)
	}

	var output []byte
	switch responseType {
	case mediatypes.HTML:
		output, err = h.bufferItemHTML(w, r, doc, docID, itemETagStrong(doc), langTag, nil)
	case mediatypes.JSON:
		output, err = bufferItemJSON(doc)
	}
	if err != nil {
		respondInternalErrorSubresource(w, r, err, docID, doc.Name)
		return
	}

	writeItem(ctx, w, docID, doc, output, http.StatusCreated, langTag, responseType)
}

func (h *handler) respondCreateErrors(w http.ResponseWriter, r *http.Request, userData FormUserData, errsField errs.Fields) {

	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)
	T := i18n.Func(langTag)
	responseType := transport.ResolveErrorContentType(r)

	var output []byte
	var err error
	switch responseType {
	case mediatypes.HTML:
		output, err = h.bufferCreateForm(ctx, newFormData(userData, &errsField, langTag))
	case mediatypes.ProblemJSON, mediatypes.JSON:
		output, err = transport.ProblemFieldError(ctx, errsField, "")
		w.Header().Set(
			"Link",
			transport.ProblemErrorsPath.
				Link(weblinking.DescribedBy).
				WithTitle(T(i18n.NavigationDocumentation)).
				String(),
		)
	case mediatypes.TextPlain:
		output = []byte(T(i18n.CommonErrorSummaryTitle) + "\n" + errsField.String())
	}
	if err != nil {
		respondInternalError(w, r, err)
		return
	}

	transport.SetErrorHeaders(w.Header())
	transport.Write(ctx, w, output, responseType, http.StatusUnprocessableEntity)
}
