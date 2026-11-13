package suggestions

import (
	"bufio"
	"customcomplete/internal/catalog"
	"customcomplete/internal/constraints"
	"customcomplete/internal/errs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/store"
	"customcomplete/internal/textprocessing"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	"database/sql"
	_ "embed"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

var (
	//go:embed suggestions.html
	tmpl           string
	supportedTypes                 = mediatypes.Types{mediatypes.HTML, mediatypes.Linkset, mediatypes.LinksetJSON}
	Path           weblinking.Path = weblinking.RootPath.Join("suggestions")
)

const (
	limitDefault           = 5
	limitMax               = 100
	maxSuggestionsInHeader = 10
	limitBytesLinkHeader   = 4 << 10 //4 KiB
	// bump for changes that should affect the ETag
	version = 1
)

type queries struct {
	autocomplete struct {
		doc []*sql.Stmt
		all []*sql.Stmt
	}
	nextWord struct {
		doc []*sql.Stmt
		all []*sql.Stmt
	}
	eTagAll *sql.Stmt
	metaDoc *sql.Stmt
}

type clientData struct {
	Text         string
	FilterTextID int64
	Limit        int64
}

type handler struct {
	queries  queries
	template map[language.Tag]*template.Template
}

func (h *handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /suggestions", redirectCanonical(h.suggest))
}

func NewHandler(db store.Read, baseTemplates *ui.BaseTemplates) (h *handler, err error) {
	h = &handler{}
	if h.queries, err = prepareQueries(db); err != nil {
		return
	}
	h.template, err = baseTemplates.BuildHTML(tmpl)
	return
}

func (h *handler) suggest(w http.ResponseWriter, r *http.Request) {
	langTag := i18n.LangTagFromContext(r.Context())
	T := i18n.Func(langTag)
	responsetype, ok := transport.NegotiateOr406(w, r,
		weblinking.New(weblinking.Up, weblinking.RootPath).WithTitle(T(i18n.NavigationHome)),
		supportedTypes,
	)
	if !ok {
		return
	}

	var suggestions []string
	var err error

	clientData := clientData{
		Text: r.URL.Query().Get("text"),
	}

	clientData.FilterTextID, err = strconv.ParseInt(r.URL.Query().Get("textID"), 10, 0)
	if err != nil {
		clientData.FilterTextID = 0
	}

	clientData.Limit, err = strconv.ParseInt(r.URL.Query().Get("limit"), 10, 0)
	if err != nil {
		clientData.Limit = limitDefault
	} else if clientData.Limit <= 0 {
		clientData.Limit = limitDefault
	} else if clientData.Limit > limitMax {
		clientData.Limit = limitMax
	}
	ctx, prefix, baseText := parseContext(clientData.Text)

	// limit + 2 prefix bounds + 1 docID
	args := make([]any, 0, len(ctx)+3)
	for _, ctxWord := range ctx {
		args = append(args, ctxWord)
	}

	var docID *catalog.DocumentID
	var ETag transport.ETagWeak
	title := T(i18n.NavigationSuggestions)
	filteredByTextID := clientData.FilterTextID != 0
	if filteredByTextID {
		var docName constraints.Label
		ETag, docID, docName, err = h.getETagDoc(clientData.FilterTextID, langTag, responsetype)
		if err != nil {
			if invalidParam, ok := errors.AsType[*errs.FieldError](err); ok {
				h.respondErrorTextID(w, r, invalidParam, clientData, title)
				return
			}
			h.respondInternalErrorQuery(w, r, err, clientData, baseText, title)
			return
		}
		args = append(args, docID)
		title = T(i18n.SuggestionsTitleFilteredTextID, docName)
	} else {
		var state string
		if err = h.queries.eTagAll.QueryRow(version).Scan(&state); err != nil {
			h.respondInternalErrorQuery(w, r, err, clientData, baseText, title)
			return
		}
		ETag = transport.GetEtagWeak([]byte(state))
	}

	if transport.ETagMatchWeak(r.Header.Values("If-None-Match"), ETag) {
		w.Header().Set("ETag", string(ETag))
		w.WriteHeader(http.StatusNotModified)
		return
	}

	hasPrefix := prefix != ""
	if hasPrefix {
		args = append(args, prefix, prefix+"\U0010ffff")
	}

	args = append(args, clientData.Limit)

	var rows *sql.Rows
	switch {
	case prefix == "" && filteredByTextID:
		rows, err = h.queries.nextWord.doc[len(ctx)].QueryContext(r.Context(), args...)
	case prefix == "" && !filteredByTextID:
		rows, err = h.queries.nextWord.all[len(ctx)].QueryContext(r.Context(), args...)
	case prefix != "" && filteredByTextID:
		rows, err = h.queries.autocomplete.doc[len(ctx)].QueryContext(r.Context(), args...)
	case prefix != "" && !filteredByTextID:
		rows, err = h.queries.autocomplete.all[len(ctx)].QueryContext(r.Context(), args...)
	}
	if err != nil {
		h.respondInternalErrorQuery(w, r, err, clientData, baseText, title)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var suggestion string
		if err := rows.Scan(&suggestion); err != nil {
			h.respondInternalErrorQuery(w, r, err, clientData, baseText, title)
			return
		}
		suggestions = append(suggestions, suggestion)
	}
	if rows.Err() != nil {
		h.respondInternalErrorQuery(w, r, err, clientData, baseText, title)
		return
	}

	if len(suggestions) == 0 && len(ctx) == 0 && prefix == "" {
		h.respondNoTextsYet(w, r, clientData, responsetype)
		return
	}

	self := weblinking.New(weblinking.Self, Path).WithQuery(r.URL.Query())
	links := append(linksGeneral(langTag), weblinking.AlternatesFor(self, langTag, responsetype, supportedTypes...)...)
	links = append(links, self)
	suggestionLink := weblinking.New(weblinking.Item, Path)
	if docID != nil {
		suggestionLink = suggestionLink.WithQueryVal("textID", strconv.FormatInt(int64(*docID), 10))
	}
	for _, suggestion := range suggestions {
		links = append(
			links,
			suggestionLink.WithQueryVal("text", baseText+suggestion+" ").
				WithTitle(suggestion),
		)
	}

	var output []byte
	switch responsetype {
	case mediatypes.HTML:
		output, err = h.bufferHTML(langTag, clientData, suggestions, false, "", pathBack(langTag, baseText, r.URL.Query()), title, false)
	case mediatypes.Linkset:
		output = []byte(links.Linkset())
	case mediatypes.LinksetJSON:
		output, err = links.MarshalJSON()
	}
	if err != nil {
		transport.InternalError(w, r, err, pathBack(langTag, baseText, r.URL.Query()))
		return
	}
	if responsetype == mediatypes.HTML {
		w.Header().Set("Link", links.StringUpTo(limitBytesLinkHeader))
	}
	w.Header().Set("ETag", string(ETag))
	transport.Write(r.Context(), w, output, responsetype, http.StatusOK)
}

func (h *handler) respondErrorTextID(w http.ResponseWriter, r *http.Request, invalidParam *errs.FieldError, uData clientData, title string) {
	langTag := i18n.LangTagFromContext(r.Context())
	T := i18n.Func(langTag)

	q := r.URL.Query()
	q.Del("textID")
	backLink := weblinking.New(weblinking.Up, Path).
		WithQuery(q).
		WithTitle(T(i18n.SuggestionsBackToLast))

	responsetype := transport.ResolveErrorContentType(r)

	links := append(linksGeneral(langTag), backLink)

	var output []byte
	var err error
	switch responsetype {
	case mediatypes.HTML:
		reason := T(i18n.SuggestionsFilterTextError) + invalidParam.Reason
		output, err = h.bufferHTML(langTag, uData, []string{}, false, reason, backLink, title, false)
	case mediatypes.JSON, mediatypes.ProblemJSON:
		output, err = transport.ProblemFieldError(r.Context(), errs.FieldErrors{*invalidParam}, T(i18n.SuggestionsFilterTextError))
		if err == nil {
			links = append(
				links,
				transport.ProblemErrorsPath.
					Link(weblinking.DescribedBy).
					WithTitle(T(i18n.NavigationDocumentation)),
			)
		}
	case mediatypes.TextPlain:
		output = []byte(T(i18n.SuggestionsFilterTextError) + ":\n" + invalidParam.Reason)
	}
	if err != nil {
		transport.InternalError(w, r, err, backLink)
		return
	}

	transport.SetErrorHeaders(w.Header())
	w.Header().Set("Link", links.String())
	transport.AddVaryOnce(w.Header(), "Accept")
	status := http.StatusUnprocessableEntity
	if invalidParam.Code == errs.CodeConflict {
		status = http.StatusConflict
	}
	transport.Write(r.Context(), w, output, responsetype, status)
}

func (h *handler) respondNoTextsYet(w http.ResponseWriter, r *http.Request, uData clientData, responsetype mediatypes.Type) {
	langTag := i18n.LangTagFromContext(r.Context())

	T := i18n.Func(langTag)

	links := append(linksGeneral(langTag), Path.Link(weblinking.Up))

	var output []byte
	var err error
	switch responsetype {
	case mediatypes.HTML:
		output, err = h.bufferHTML(langTag, uData, []string{}, false, "", Path.Link(weblinking.Up), T(i18n.NavigationSuggestions), true)
	case mediatypes.Linkset:
		output = []byte(links.Linkset())
	case mediatypes.LinksetJSON:
		output, err = links.MarshalJSON()
	}
	if err != nil {
		transport.InternalError(w, r, err, Path.Link(weblinking.Up))
		return
	}
	w.Header().Set("Link", links.String())
	transport.Write(r.Context(), w, output, responsetype, http.StatusOK)

}

func (h *handler) respondInternalErrorQuery(w http.ResponseWriter, r *http.Request, errQuery error, uData clientData, baseText string, title string) {
	slog.WarnContext(r.Context(), "Internal Error: "+errQuery.Error())

	langTag := i18n.LangTagFromContext(r.Context())
	T := i18n.Func(langTag)

	backLink := pathBack(langTag, baseText, r.URL.Query())

	responsetype := transport.ResolveErrorContentType(r)

	var output []byte
	var err error
	switch responsetype {
	case mediatypes.HTML:
		output, err = h.bufferHTML(langTag, uData, []string{}, true, "", backLink, title, false)
	case mediatypes.JSON, mediatypes.ProblemJSON:
		output, err = transport.ProblemGeneric(T(i18n.SuggestionsError))
	case mediatypes.TextPlain:
		output = []byte(T(i18n.SuggestionsError))
	}
	if err != nil {
		transport.InternalError(w, r, err, backLink)
		return
	}
	transport.SetErrorHeaders(w.Header())
	w.Header().Set("Link", append(linksGeneral(langTag), backLink).String())
	transport.Write(r.Context(), w, output, responsetype, http.StatusInternalServerError)
}

func parseContext(text string) (context []string, prefix string, baseText string) {

	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Split(textprocessing.ScanNormalizedWords)

	var words []string
	var lastRaw string
	for scanner.Scan() {
		lastRaw = scanner.Text()
		words = append(words, textprocessing.NFKCFold(scanner.Text()))
	}
	if len(words) == 0 {
		return nil, "", text
	}

	if strings.HasSuffix(text, lastRaw) {
		prefix = words[len(words)-1]
		words = words[:len(words)-1]
		baseText = text[:len(text)-len(lastRaw)]
	} else {
		baseText = text
	}

	if len(words) > maxContextWords {
		words = words[len(words)-maxContextWords:]
	}

	return words, prefix, baseText
}

func (h *handler) getETagDoc(rawID int64, langTag language.Tag, responseType mediatypes.Type) (ETag transport.ETagWeak, docID *catalog.DocumentID, docName constraints.Label, err error) {
	var ok bool
	var id catalog.DocumentID
	T := i18n.Func(langTag)
	id, ok = catalog.ParseDocumentIDFromInt64(rawID)
	if !ok {
		return "", nil, "", new(errs.InvalidParamFromCodedError("textID", errs.NotFound(), T(i18n.SuggestionsFilterTextError)+" "+T(i18n.SuggestionsFilterTextNotFound)))
	}
	docID = &id
	var state string
	var status catalog.SourceStatus
	if err = h.queries.metaDoc.QueryRow(version, docID).Scan(&state, &status, &docName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, "", new(errs.InvalidParamFromCodedError("textID", errs.NotFound(), T(i18n.SuggestionsFilterTextError)+" "+T(i18n.SuggestionsFilterTextNotFound)))
		}
		return
	}
	var reason string
	switch status {
	case catalog.StatusComplete:
		return transport.GetEtagWeak([]byte([]byte(state + "|" + responseType.String() + "|" + langTag.String()))), docID, docName, nil
	case catalog.StatusFailed:
		reason = T(i18n.SuggestionsFilterTextProcessingFailed)
	case catalog.StatusPending, catalog.StatusProcessing:
		reason = T(i18n.SuggestionsFilterTextProcessingIncomplete)
	}

	return "", nil, "", new(errs.InvalidParamFromCodedError("textID", errs.Conflict(catalog.StatusComplete, status), reason))
}

func (h *handler) bufferHTML(langTag language.Tag, uData clientData, suggestions []string, hasErrorQuery bool, errMessage string, linkBack weblinking.Link, title string, NoTexts bool) ([]byte, error) {

	return ui.BufferHTML(h.template[langTag], struct {
		Base ui.BaseDataHTML
		clientData
		Suggestions   []string
		InternalError bool
		Error         string
		LinkBack      weblinking.Link
		NoTexts       bool
	}{
		Base: ui.BaseDataHTML{
			Title: title,
			Lang:  langTag,
			Breadcrumbs: weblinking.Links{
				weblinking.New(weblinking.Up, weblinking.RootPath).
					WithTitle(i18n.Func(langTag)(i18n.NavigationHome)),
			},
		},
		clientData:    uData,
		Suggestions:   suggestions,
		InternalError: hasErrorQuery,
		Error:         errMessage,
		LinkBack:      linkBack,
		NoTexts:       NoTexts,
	})
}

func linksGeneral(langTag language.Tag) weblinking.Links {
	T := i18n.Func(langTag)
	return weblinking.Links{
		weblinking.New(weblinking.Up, weblinking.RootPath).WithTitle(T(i18n.NavigationHome)),
	}
}

func pathBack(langTag language.Tag, baseText string, query url.Values) weblinking.Link {
	out := weblinking.New(weblinking.Up, Path).
		WithQueryVal("text", baseText).
		WithTitle(i18n.Translate(langTag, i18n.SuggestionsBackToLast))
	if query.Has("textID") {
		out = out.WithQueryVal("textID", query.Get("textID"))
	}
	return out
}
