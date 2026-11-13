package texts

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"encoding/json"
	"net/http"

	"golang.org/x/text/language"
)

var (
	//go:embed collection.html
	collectionTemplate       string
	collectionSupportedTypes mediatypes.Types = mediatypes.Types{
		mediatypes.HTML, mediatypes.LinksetJSON, mediatypes.Linkset,
	}
	CollectionPath     weblinking.Path = weblinking.RootPath.Join("texts")
	CollectionSpecPath weblinking.Path = specPath.Join("collection")
	CollectionDocPath  weblinking.Path = docPath.Join("collection")
)

func collectionLink(langTag language.Tag) weblinking.Link {
	return CollectionPath.Link(weblinking.Collection).
		WithTitle(i18n.Translate(langTag, i18n.NavigationTextCollection))
}

func (h *handler) registerCollection(mux *http.ServeMux) {
	mux.Handle("GET "+CollectionPath.String(), transport.NoBody(h.collection))
	mux.Handle("OPTIONS "+CollectionPath.String(), transport.NoBody(h.collectionOptions))
}

func (h *handler) initCollection(base *ui.BaseTemplates) (err error) {
	h.collectionTemplate, err = base.BuildHTML(collectionTemplate)
	return
}

func (h *handler) collection(w http.ResponseWriter, r *http.Request) {

	langTag := i18n.LangTagFromContext(r.Context())
	T := i18n.Func(langTag)
	homeLink := weblinking.RootLink(langTag)

	responseType, ok := transport.NegotiateOr406(w, r, homeLink, collectionSupportedTypes)
	if !ok {
		return
	}

	refs, err := h.service.GetDocumentRefs()
	if err != nil {
		transport.InternalErrorGeneric(w, r, err)
		return
	}

	links := make(weblinking.Links, 0, len(refs))
	for _, ref := range refs {
		links = append(links, weblinking.New(weblinking.Item, TextsItemPath(ref.ID)).WithTitle(string(ref.Name)))
	}

	createFormLink := weblinking.New(weblinking.CreateForm, CreateFormPath).WithTitle(T(i18n.NavigationTextCreateForm))
	if responseType != mediatypes.HTML {
		createFormLink = createFormLink.WithType(mediatypes.HTML)
	}

	headerLinks := append(weblinking.Links{
		homeLink,
		weblinking.New(weblinking.Self, CollectionPath),
		createFormLink,
		weblinking.New(weblinking.DescribedBy, CollectionSpecPath).
			WithTitle(T(i18n.NavigationDocumentation)).
			WithType(mediatypes.OpenAPIJSON),
		weblinking.New(weblinking.DescribedBy, CollectionDocPath).
			WithTitle(T(i18n.NavigationDocumentation)).
			WithType(mediatypes.HTML),
	}, weblinking.Alternates(CollectionPath, langTag, responseType, collectionSupportedTypes...)...)

	if responseType != mediatypes.HTML {
		links = append(links, headerLinks...)
	}

	var output []byte
	switch responseType {
	case mediatypes.HTML:
		output, err = ui.BufferHTML(h.collectionTemplate[langTag], struct {
			Base       ui.BaseDataHTML
			Items      weblinking.Links
			Flash      string
			CreateForm weblinking.Link
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.NavigationTextCollection),
				MetaDescription: T(i18n.TextCollectionMetaDescription),
				Lang:            langTag,
				Breadcrumbs:     weblinking.Links{homeLink},
			},
			Items:      links,
			Flash:      consumeFlash(w, r, CollectionPath, flashDeleted),
			CreateForm: createFormLink,
		})
	case mediatypes.LinksetJSON:
		output, err = json.Marshal(links)
	case mediatypes.Linkset:
		output = []byte(links.Linkset())
	}
	if err != nil {
		transport.InternalErrorGeneric(w, r, err)
		return
	}

	w.Header().Set("Accept-Post", postContentType.String())
	w.Header().Set("Allow", "GET, HEAD, POST, OPTIONS")
	transport.SetNoCache(w.Header())
	etagWeak := transport.GetEtagWeak(output)
	w.Header().Set("ETag", string(etagWeak))
	w.Header().Set("Link", headerLinks.String())
	if transport.ETagMatchWeak(r.Header.Values("If-None-Match"), etagWeak) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	transport.Write(r.Context(), w, output, responseType, http.StatusOK)
}

func (h *handler) collectionOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Del("Content-Language")

	w.Header().Set("Allow", "GET, HEAD, POST, OPTIONS")
	w.Header().Set("Accept-Post", postContentType.String())

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Link, Allow, Accept-Post, Location, Preference-Applied, Content-Location, Set-Cookie, Vary")

	langTag := i18n.LangTagFromContext(r.Context())
	T := i18n.Func(langTag)
	w.Header().Set("Link", append(
		weblinking.Links{
			weblinking.RootLink(langTag),
			weblinking.New(weblinking.Self, CollectionPath).
				WithTitle(T(i18n.NavigationTextCollection)).
				WithTitleLang(langTag),
			weblinking.New(weblinking.CreateForm, CreateFormPath).
				WithTitle(T(i18n.NavigationTextCreateForm)).
				WithType(mediatypes.HTML).
				WithTitleLang(langTag),
			weblinking.New(weblinking.DescribedBy, CollectionDocPath).
				WithTitle(T(i18n.NavigationDocumentation)).WithType(mediatypes.HTML).
				WithTitleLang(langTag),
			weblinking.New(weblinking.DescribedBy, CollectionSpecPath).
				WithTitle(T(i18n.NavigationDocumentation)).WithType(mediatypes.OpenAPIJSON).
				WithTitleLang(langTag),
		},
		weblinking.AllAlternates(CollectionPath, collectionSupportedTypes...)...,
	).String())

	w.WriteHeader(http.StatusNoContent)
}
