package transport

import (
	"context"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	"fmt"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

func adjustServerOrder(supported mediatypes.Types, sent mediatypes.Type) (reordered mediatypes.Types) {
	lowPrio := mediatypes.Types{}
	for _, support := range supported {
		if sent.SharesSyntax(support) {
			reordered = append(reordered, support)
		} else {
			lowPrio = append(lowPrio, support)
		}
	}
	return append(reordered, lowPrio...)
}

func ResolveErrorContentType(r *http.Request) mediatypes.Type {

	errorMediaTypes := mediatypes.Types{
		mediatypes.HTML,
		mediatypes.ProblemJSON,
		mediatypes.JSON,
		mediatypes.TextPlain,
	}

	requestType, errRequestType := mediatypes.ParseMediaType(r.Header.Get("Content-Type"))
	hasRequestType := errRequestType == nil
	if hasRequestType {
		errorMediaTypes = adjustServerOrder(errorMediaTypes, requestType)
	}
	requested := mediatypes.ParseAccept(r.Header.Values("Accept"))

	match, ok := errorMediaTypes.Negotiate(requested)
	if ok {
		return match
	}

	if hasRequestType && requestType.SharesSyntax(errorMediaTypes[0]) {
		return errorMediaTypes[0]
	}

	return mediatypes.TextPlain
}

func negotiate(r *http.Request, supported mediatypes.Types) (mediatypes.Type, bool) {
	requested := mediatypes.ParseAccept(r.Header.Values("Accept"))
	if requestType, err := mediatypes.ParseMediaType(r.Header.Get("Content-Type")); err == nil {
		supported = adjustServerOrder(supported, requestType)
	}
	return supported.Negotiate(requested)
}

func NegotiateOr406(w http.ResponseWriter, r *http.Request, escape weblinking.Link, supported mediatypes.Types) (mediatypes.Type, bool) {
	AddVaryOnce(w.Header(), "Accept")
	match, ok := negotiate(r, supported)
	if !ok {
		NotAcceptable(w, r, escape, supported)
		return mediatypes.Type{}, false
	}
	return match, true
}

func NegotiateOrFallback(r *http.Request, supported mediatypes.Types, fallback mediatypes.Type) mediatypes.Type {
	requested := mediatypes.ParseAccept(r.Header.Values("Accept"))
	requestType, err := mediatypes.ParseMediaType(r.Header.Get("Content-Type"))
	hasRequestType := err == nil
	if hasRequestType {
		supported = adjustServerOrder(supported, requestType)
	}
	negotiated, ok := supported.Negotiate(requested)
	if ok {
		return negotiated
	}
	if hasRequestType && requestType.SharesSyntax(supported[0]) {
		return supported[0]
	}
	return fallback
}

func RequireAccepts(mediaType mediatypes.Type, w http.ResponseWriter, r *http.Request, escape weblinking.Link) bool {
	_, ok := NegotiateOr406(w, r, escape, mediatypes.Types{mediaType})
	return ok
}

// decides if the request should follow the PRG Pattern
func IsFullDocumentNavigation(r *http.Request) bool {

	// only relevant for state-changing requests
	if r.Method != http.MethodPost {
		return false
	}

	fetchDestination := r.Header.Get("Sec-Fetch-Dest")
	fetchMode := r.Header.Get("Sec-Fetch-Mode")
	// fetch metadata request headers indicate document navigation
	if strings.EqualFold(fetchDestination, "document") || strings.EqualFold(fetchMode, "navigate") {
		return true
	}
	// fetch metadata request headers exist but dont indicate document navigation
	if fetchDestination != "" || fetchMode != "" || r.Header.Get("Sec-Fetch-Site") != "" {
		return false
	}

	// document navigation doesnt send "Prefer"-header
	if r.Header.Get("Prefer") != "" {
		return false
	}

	// document navigation doesnt send "If-Match"-header
	if r.Header.Get("If-Match") != "" {
		return false
	}

	// html form submission is either url-encoded or multipart form-data
	mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mediaType != "application/x-www-form-urlencoded" && mediaType != "multipart/form-data" {
		return false
	}

	// explicit custom header setting via js
	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		return false
	}

	// client explicitly (not just via wildcard) accepts problem json
	requested := mediatypes.ParseAccept(r.Header.Values("Accept"))
	if requested.ExplicitlyAccepts(mediatypes.ProblemJSON) {
		return false
	}

	// client doesnt prioritize html
	match, ok := mediatypes.Types{mediatypes.JSON, mediatypes.TextPlain, mediatypes.HTML}.Negotiate(requested)
	if !ok || match != mediatypes.HTML {
		return false
	}

	// assume full document navigation as fallback
	return true
}

func SetErrorHeaders(header http.Header) {
	SetDontCache(header)
	AddVaryOnce(header, "Accept")
	AddVaryOnce(header, "Accept-Language")
	ClearPreferences(header)
}

func InternalError(w http.ResponseWriter, r *http.Request, err error, escape weblinking.Link) {
	slog.WarnContext(r.Context(), "Internal Error", slog.Any("error", err.Error()))

	T := i18n.TranslateFuncFromContext(r.Context())

	writeError(w, r,
		http.StatusInternalServerError,
		escape,
		fmt.Sprintf("%s. %s", T(i18n.CommonErrorService), T(i18n.HTTPErrorInternalServerErrorTryAgain)),
		func(langTag language.Tag) ([]byte, error) {
			return ui.InternalServerErrorHTML(langTag, escape)
		})
}

func InternalErrorGeneric(w http.ResponseWriter, r *http.Request, err error) {
	InternalError(w, r, err, weblinking.New(weblinking.Up, weblinking.RootPath))
}

// requires negotiated type to be html
func SeeOther(w http.ResponseWriter, r *http.Request, location weblinking.Path, message string) {

	ctx := r.Context()
	langtag := i18n.LangTagFromContext(ctx)

	w.Header().Set("Location", location.String())

	var err error
	var body []byte

	body, err = ui.SeeOtherHTML(langtag, message, location)
	if err != nil {
		log.Println("failed to buffer template 303")
		http.Redirect(w, r, location.String(), http.StatusSeeOther)
		return
	}

	Write(ctx, w, body, mediatypes.HTML, http.StatusSeeOther)
}

func BadRequest(w http.ResponseWriter, r *http.Request, detail string, escape weblinking.Link) {
	writeError(w, r,
		http.StatusBadRequest,
		escape,
		detail,
		func(langTag language.Tag) ([]byte, error) {
			return ui.BadRequestHTML(langTag, escape, detail)
		})
}

func UnsupportedMediatype(w http.ResponseWriter, r *http.Request, escape weblinking.Link, sent mediatypes.Type, supported mediatypes.Types) {
	switch r.Method {
	case http.MethodPost:
		w.Header().Set("Accept-Post", supported.String())
	case http.MethodPatch:
		w.Header().Add("Accept-Patch", supported.String())
	}

	SetDontCache(w.Header())

	responsetype := ResolveErrorContentType(r)

	w.Header().Set("Content-Type", responsetype.String())

	T := i18n.TranslateFuncFromContext(r.Context())

	writeError(w, r,
		http.StatusUnsupportedMediaType,
		escape,
		T(i18n.HTTPErrorUnsupportedMediaTypeDescription, sent, supported),
		func(langTag language.Tag) ([]byte, error) {
			return ui.UnsupportedMediaTypeHTML(langTag, escape, sent, supported)
		})
}

func NotAcceptable(w http.ResponseWriter, r *http.Request, escape weblinking.Link, supported mediatypes.Types) {
	T := i18n.TranslateFuncFromContext(r.Context())
	writeError(w, r,
		http.StatusNotAcceptable,
		escape,
		T(i18n.HTTPErrorNotAcceptableDescription)+" "+T(i18n.HTTPErrorNotAcceptableAvailable, supported),
		func(langTag language.Tag) ([]byte, error) {
			return ui.NotAcceptableHTML(langTag, escape, supported)
		})
}

func NotFound(w http.ResponseWriter, r *http.Request, escape weblinking.Link) {
	T := i18n.TranslateFuncFromContext(r.Context())
	writeError(w, r,
		http.StatusNotFound,
		escape,
		T(i18n.HTTPErrorNotFound)+"\n"+T(i18n.HTTPErrorNotFoundDescription),
		func(langTag language.Tag) ([]byte, error) {
			return ui.NotFoundHTML(langTag, escape)
		})
}

func ContentTooLarge(w http.ResponseWriter, r *http.Request, limiFmt string, escape weblinking.Link) {
	T := i18n.TranslateFuncFromContext(r.Context())

	writeError(w, r,
		http.StatusRequestEntityTooLarge,
		escape,
		T(i18n.HTTPErrorContentTooLargeDescription, limiFmt),
		func(langTag language.Tag) ([]byte, error) {
			return ui.ContentTooLargeHTML(langTag, limiFmt, escape)
		})
}

func BodyNotAllowed(w http.ResponseWriter, r *http.Request, langTag language.Tag) {
	T := i18n.TranslateFuncFromContext(r.Context())

	escapeHome := weblinking.New(weblinking.Up, weblinking.RootPath).WithTitle(T(i18n.NavigationHome))

	writeError(w, r,
		http.StatusBadRequest,
		escapeHome,
		T(i18n.HTTPErrorRequestBodyNotAllowed),
		func(langTag language.Tag) ([]byte, error) {
			return ui.RequestBodyNotAllowedHTML(langTag, escapeHome)
		})
}

func writeError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	escape weblinking.Link,
	detail string,
	htmlFn func(language.Tag) ([]byte, error),
) {
	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)
	responseType := ResolveErrorContentType(r)

	h := w.Header()
	SetErrorHeaders(h)
	if !escape.IsZero() {
		h.Add("Link", escape.String())
	}

	var body []byte
	var err error
	switch responseType {
	case mediatypes.HTML:
		body, err = htmlFn(langTag)
	case mediatypes.ProblemJSON, mediatypes.JSON:
		body, err = ProblemGeneric(detail)
	default:
		responseType = mediatypes.TextPlain
		body = []byte(detail)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to render error response", slog.Int("status", status), slog.Any("err", err))
		http.Error(w, detail, status)
		return
	}

	Write(ctx, w, body, responseType, status)
}

func Write(ctx context.Context, w http.ResponseWriter, output []byte, responseType mediatypes.Type, status int) {
	w.Header().Set("Content-Type", responseType.String())
	w.Header().Set("Content-Length", strconv.Itoa(len(output)))
	w.WriteHeader(status)
	if _, err := w.Write(output); err != nil {
		slog.WarnContext(ctx, "could not write")
	}
}

func WriteNoBody(ctx context.Context, w http.ResponseWriter, status int) {
	w.Header().Del("Content-Type")
	w.Header().Del("Content-Language")
	w.WriteHeader(status)
}
