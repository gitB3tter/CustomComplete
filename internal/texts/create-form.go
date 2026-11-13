package texts

import (
	"context"
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
	CreateFormPath     weblinking.Path = CollectionPath.Join("new")
)

func (h *handler) registerCreateForm(mux *http.ServeMux) {
	mux.Handle("GET "+CreateFormPath.String(), transport.NoBody(h.createForm))
	mux.Handle("OPTIONS "+CreateFormPath.String(), transport.NoBody(h.createFormOptions))
}

func (h *handler) initCreateForm(base *ui.BaseTemplates) (err error) {
	h.createFormTemplate, err = base.BuildHTML(createFormTemplate)
	return
}

func (h *handler) createForm(w http.ResponseWriter, r *http.Request) {

	langTag := i18n.LangTagFromContext(r.Context())

	if !transport.RequireAccepts(mediatypes.HTML, w, r, collectionLink(langTag)) {
		return
	}

	userData := FormUserData{}
	if text := r.URL.Query().Get("text"); text != "" {
		userData.Source = textField
		userData.Text = text
	}

	output, err := h.bufferCreateForm(r.Context(), newFormData(userData, nil, langTag))
	if err != nil {
		respondInternalError(w, r, err)
		return
	}

	T := i18n.Func(langTag)

	w.Header().Set("Link", append(
		weblinking.Links{
			CollectionPath.Link(weblinking.Up).
				WithTitle(T(i18n.NavigationTextCollection)),
			collectionLink(langTag),
			CreateFormPath.Link(weblinking.Self).
				WithTitle(T(i18n.NavigationTextCreateForm)),
			CollectionSpecPath.Link(weblinking.DescribedBy).
				WithTitle(T(i18n.NavigationDocumentation)).
				WithType(mediatypes.OpenAPIJSON),
			CollectionDocPath.Link(weblinking.DescribedBy).
				WithTitle(T(i18n.NavigationDocumentation)).
				WithType(mediatypes.HTML),
		},
		weblinking.Alternates(CreateFormPath, langTag, mediatypes.HTML)...,
	).String())

	etagWeak := transport.GetEtagWeak(output)
	w.Header().Set("ETag", string(etagWeak))
	if transport.ETagMatchWeak(r.Header.Values("If-None-Match"), etagWeak) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Allow", "GET, HEAD, OPTIONS")
	transport.Write(r.Context(), w, output, mediatypes.HTML, http.StatusOK)
}

func (h *handler) bufferCreateForm(ctx context.Context, formdata formData) ([]byte, error) {

	langTag := i18n.LangTagFromContext(ctx)
	T := i18n.Func(langTag)

	var output []byte
	var err error

	if formdata.Source != textField {
		formdata.Text = ""
	}

	if output, err = ui.BufferHTML(h.createFormTemplate[langTag], struct {
		Base ui.BaseDataHTML
		formData
	}{
		Base: ui.BaseDataHTML{
			Title:           T(i18n.NavigationTextCreateForm),
			MetaDescription: T(i18n.TextCreateFormMetaDescription),
			Lang:            langTag,
			NoIndex:         true,
			HasError:        len(formdata.Errors) > 0,
			Back: CollectionPath.Link(weblinking.Collection).
				WithTitle(T(i18n.NavigationTextCollection)),
		},
		formData: formdata,
	}); err != nil {
		return nil, err
	}

	return output, nil
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
			weblinking.New(weblinking.Up, CollectionPath).
				WithTitle(T(i18n.NavigationTextCollection)),
			CollectionPath.Link(weblinking.Collection).
				WithTitle(T(i18n.NavigationTextCollection)),
			weblinking.New(weblinking.Self, CreateFormPath).
				WithTitle(T(i18n.NavigationTextCreateForm)).
				WithType(mediatypes.HTML),
			weblinking.New(weblinking.DescribedBy, CollectionSpecPath).
				WithTitle(T(i18n.NavigationDocumentation)).
				WithType(mediatypes.OpenAPIJSON),
			weblinking.New(weblinking.DescribedBy, CollectionDocPath).
				WithTitle(T(i18n.NavigationDocumentation)).
				WithType(mediatypes.HTML),
		},
		weblinking.AllAlternates(CreateFormPath, mediatypes.HTML)...,
	).String())

	w.WriteHeader(http.StatusNoContent)
}
