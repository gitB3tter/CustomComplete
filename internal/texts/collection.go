package texts

import (
	"customcomplete/internal/docs"
	"customcomplete/internal/home"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"encoding/json"
	"net/http"
)

var (
	//go:embed collection.html
	collectionTemplate       string
	supportedTypesCollection mediatypes.Types = mediatypes.Types{mediatypes.HTML, mediatypes.LinksetJSON, mediatypes.Linkset}
)

func (h *handler) registerCollection(mux *http.ServeMux) {
	mux.Handle("GET /texts", transport.NoBody(h.collection))
	mux.Handle("OPTIONS /texts", transport.NoBody(h.collectionOptions))
}

func (h *handler) initCollection(base *ui.BaseTemplates) (err error) {
	h.collectionTemplate, err = base.BuildHTML(collectionTemplate)
	return
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
			weblinking.New(weblinking.Up, home.Path()).WithTitle(T(i18n.NavigationHome)).WithTitleLang(langTag),
			weblinking.New(weblinking.Self, collectionPath()).WithTitle(T(i18n.NavigationTextCollection)).WithTitleLang(langTag),
			weblinking.New(weblinking.CreateForm, collectionPath().JoinPath("new")).WithTitle(T(i18n.NavigationTextCreateForm)).WithType(mediatypes.HTML).WithTitleLang(langTag),
			weblinking.New(weblinking.DescribedBy, docs.GetPath(docs.SpecsTextsCollection)).WithTitle(T(i18n.DocsTitlePrefix)).WithType(mediatypes.OpenAPIJSON).WithTitleLang(langTag),
			weblinking.New(weblinking.DescribedBy, docs.GetPath(docs.TextsCollection)).WithTitle(T(i18n.DocsTitlePrefix)).WithType(mediatypes.HTML).WithTitleLang(langTag),
		},
		weblinking.AllAlternates(collectionPath(), supportedTypesCollection...)...,
	).String())

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) collection(w http.ResponseWriter, r *http.Request) {

	lang := i18n.LangTagFromContext(r.Context())
	T := i18n.Func(lang)

	negotiated, ok := transport.NegotiateOr406(w, r, weblinking.New(weblinking.Up, home.Path()).WithTitle(T(i18n.NavigationHome)), supportedTypesCollection)
	if !ok {
		return
	}

	refs, err := h.service.GetRefs()
	if err != nil {
		transport.InternalErrorGeneric(w, r, err)
		return
	}

	links := make(weblinking.Links, 0, len(refs))
	for _, ref := range refs {
		links = append(links, weblinking.New(weblinking.Item, itemPath(ref.ID)).WithTitle(string(ref.Name)))
	}

	home := weblinking.New(weblinking.Up, home.Path()).WithTitle(T(i18n.NavigationHome))
	createForm := weblinking.New(weblinking.CreateForm, collectionPath().JoinPath("new")).WithTitle(T(i18n.NavigationTextCreateForm))
	if negotiated != mediatypes.HTML {
		createForm = createForm.WithType(mediatypes.HTML)
	}

	headerLinks := append(weblinking.Links{
		home,
		weblinking.New(weblinking.Self, collectionPath()),
		createForm,
		weblinking.New(weblinking.DescribedBy, docs.GetPath(docs.SpecsTextsCollection)).WithTitle(T(i18n.DocsTitlePrefix)).WithType(mediatypes.OpenAPIJSON),
		weblinking.New(weblinking.DescribedBy, docs.GetPath(docs.TextsCollection)).WithTitle(T(i18n.DocsTitlePrefix)).WithType(mediatypes.HTML),
	}, weblinking.Alternates(collectionPath(), lang, negotiated, supportedTypesCollection...)...)

	if negotiated != mediatypes.HTML {
		links = append(links, headerLinks...)
	}

	var output []byte
	switch negotiated {
	case mediatypes.HTML:

		output, err = ui.BufferHTML(h.collectionTemplate[lang], struct {
			Base       ui.BaseDataHTML
			Items      weblinking.Links
			Flash      string
			CreateForm weblinking.Link
		}{
			Base: ui.BaseDataHTML{
				Title:           T(i18n.NavigationTextCollection),
				MetaDescription: T(i18n.TextCollectionMetaDescription),
				Lang:            lang,
				Breadcrumbs:     weblinking.Links{home},
			},
			Items:      links,
			Flash:      consumeFlash(w, r, collectionPath(), flashDeleted),
			CreateForm: createForm,
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

	w.Header().Set("Link", headerLinks.String())

	w.Header().Set("Content-Type", negotiated.String())

	w.Header().Set("Allow", "GET, HEAD, POST, OPTIONS")
	w.Header().Set("Accept-Post", postContentType.String())

	transport.SetNoCache(w.Header())
	ETag := transport.GetEtagWeak(output)
	w.Header().Set("ETag", string(ETag))
	if transport.ETagMatchWeak(r.Header.Values("If-None-Match"), ETag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	transport.Write(r.Context(), w, output)
}
