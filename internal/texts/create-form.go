package texts

import (
	"context"
	"customcomplete/internal/docs"
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/transport"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"net/http"
)

var (
	//go:embed create-form.html
	createFormTemplate string
)

func (h *handler) registerNew(mux *http.ServeMux) {

	mux.Handle("GET /texts/new", transport.NoBody(h.createForm))
	mux.Handle("OPTIONS /texts/new", transport.NoBody(h.createFormOptions))
}

func (h *handler) initCreateForm(base *ui.BaseTemplates) (err error) {
	h.createFormTemplate, err = base.BuildHTML(createFormTemplate)
	return
}

func (h *handler) createFormOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Del("Content-Language")

	w.Header().Set("Allow", "GET, HEAD, OPTIONS")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Link, Allow, Vary")

	T := i18n.TranslateFuncFromContext(r.Context())
	w.Header().Set("Link", append(
		weblinking.Links{
			weblinking.New(weblinking.Up, collectionPath()).WithTitle(T(i18n.NavigationTextCollection)),
			weblinking.New(weblinking.Collection, collectionPath()).WithTitle(T(i18n.NavigationTextCollection)),
			weblinking.New(weblinking.Self, collectionPath().JoinPath("new")).WithTitle(T(i18n.NavigationTextCreateForm)).WithType(mediatypes.HTML),
			weblinking.New(weblinking.DescribedBy, docs.GetPath(docs.SpecsTextsCollection)).WithTitle(T(i18n.DocsTitlePrefix)).WithType(mediatypes.OpenAPIJSON),
			weblinking.New(weblinking.DescribedBy, docs.GetPath(docs.TextsCollection)).WithTitle(T(i18n.DocsTitlePrefix)).WithType(mediatypes.HTML),
		},
		weblinking.AllAlternates(collectionPath().JoinPath("new"), mediatypes.HTML)...,
	).String())

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) createForm(w http.ResponseWriter, r *http.Request) {

	if !transport.RequireAccepts(mediatypes.HTML, w, r, weblinking.New(weblinking.Collection, collectionPath())) {
		return
	}

	ctx := r.Context()
	langTag := i18n.LangTagFromContext(ctx)

	initialData := FormUserData{}
	if prefillText := r.URL.Query().Get("text"); prefillText != "" {
		initialData.Source = fieldTextName
		initialData.Text = prefillText
	}

	output, err := h.bufferCreateForm(ctx, newFormData(initialData, nil, langTag))
	if err != nil {
		respondInternalError(w, r, err)
		return
	}

	T := i18n.Func(langTag)

	w.Header().Set("Link", append(
		weblinking.Links{
			weblinking.New(weblinking.Up, collectionPath()).WithTitle(T(i18n.NavigationTextCollection)),
			weblinking.New(weblinking.Collection, collectionPath()).WithTitle(T(i18n.NavigationTextCollection)),
			weblinking.New(weblinking.Self, collectionPath().JoinPath("new")).WithTitle(T(i18n.NavigationTextCreateForm)),
			weblinking.New(weblinking.DescribedBy, docs.GetPath(docs.SpecsTextsCollection)).WithTitle(T(i18n.DocsTitlePrefix)).WithType(mediatypes.OpenAPIJSON),
			weblinking.New(weblinking.DescribedBy, docs.GetPath(docs.TextsCollection)).WithTitle(T(i18n.DocsTitlePrefix)).WithType(mediatypes.HTML),
		},
		weblinking.Alternates(collectionPath().JoinPath("new"), langTag, mediatypes.HTML)...,
	).String())

	ETagWeak := transport.GetEtagWeak(output)
	w.Header().Set("ETag", string(ETagWeak))
	if transport.ETagMatchWeak(r.Header.Values("If-None-Match"), ETagWeak) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", mediatypes.HTML.String())
	w.Header().Set("Allow", "GET, HEAD, OPTIONS")

	transport.Write(ctx, w, output)
}

func (h *handler) bufferCreateForm(ctx context.Context, formdata formData) ([]byte, error) {

	lang := i18n.LangTagFromContext(ctx)
	T := i18n.Func(lang)

	var buffer []byte
	var err error

	if formdata.Source != fieldTextName {
		formdata.Text = ""
	}

	if buffer, err = ui.BufferHTML(h.createFormTemplate[lang], struct {
		Base ui.BaseDataHTML
		formData
	}{
		Base: ui.BaseDataHTML{
			Title:           T(i18n.NavigationTextCreateForm),
			MetaDescription: T(i18n.TextCreateFormMetaDescription),
			Lang:            lang,
			NoIndex:         true,
			HasError:        len(formdata.Errors) > 0,
			Back:            weblinking.New(weblinking.Collection, collectionPath()).WithTitle(T(i18n.NavigationTextCollection)),
		},
		formData: formdata,
	}); err != nil {
		return nil, err
	}

	return buffer, nil
}
