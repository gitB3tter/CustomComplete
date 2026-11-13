package texts

import (
	"customcomplete/internal/catalog"
	"customcomplete/internal/i18n"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"time"

	"golang.org/x/text/language"
)

var (
	//reused pieces
	//go:embed partials/texts.html
	tmplTexts string
)

type flashName string

const (
	flashCreated                  flashName = "created"
	flashUpdated                  flashName = "updated"
	flashDeleted                  flashName = "deleted"
	flashDeletePreconditionFailed flashName = "delete-precondition-failed"
	flashDeleteNoPrecondition     flashName = "delete-no-precondition"
)

type handler struct {
	service            *catalog.Service
	collectionTemplate map[language.Tag]*template.Template
	createFormTemplate map[language.Tag]*template.Template
	itemTemplate       map[language.Tag]*template.Template
	statusTemplate     map[language.Tag]*template.Template
	editFormTemplate   map[language.Tag]*template.Template
	deleteFormTemplate map[language.Tag]*template.Template
}

func (h *handler) Register(mux *http.ServeMux) {

	h.registerCollection(mux)
	h.registerCreateForm(mux)
	h.registerCreate(mux)
	h.registerItem(mux)
	h.registerEditForm(mux)
	h.registerUpdate(mux)
	h.registerDelete(mux)
	h.registerDeleteForm(mux)
	h.registerStatus(mux)
	h.registerFile(mux)

}

func NewHandler(service *catalog.Service, baseTemplates *ui.BaseTemplates) (h *handler, err error) {

	h = &handler{}

	if err := baseTemplates.WithHTML(tmplTexts); err != nil {
		return nil, err
	}

	if err = h.initCollection(baseTemplates); err != nil {
		return
	}
	if err = h.initCreateForm(baseTemplates); err != nil {
		return
	}
	if err = h.initItem(baseTemplates); err != nil {
		return
	}
	if err = h.initStatus(baseTemplates); err != nil {
		return nil, err
	}
	if err = h.initEditForm(baseTemplates); err != nil {
		return nil, err
	}
	if err = h.initDeleteForm(baseTemplates); err != nil {
		return nil, err
	}

	h.service = service

	return h, nil
}

func consumeFlash(w http.ResponseWriter, r *http.Request, path weblinking.Path, names ...flashName) (flash string) {
	T := i18n.TranslateFuncFromContext(r.Context())
	for _, cookie := range r.Cookies() {
		if slices.ContainsFunc(names, func(n flashName) bool { return string(n) == cookie.Name }) {
			switch cookie.Name {
			case string(flashCreated):
				flash = T(i18n.TextCreateFlashSuccess)
			case string(flashUpdated):
				flash = T(i18n.TextUpdateFlashSuccess)
			case string(flashDeleted):
				name, err := url.QueryUnescape(cookie.Value)
				if err != nil {
					slog.WarnContext(r.Context(), fmt.Sprintf("Cookie Value %s of cookie %s not unescapeable", cookie.Value, cookie.Name))
					name = ""
				}
				flash = T(i18n.TextDeleteFlashSuccess, name)
			case string(flashDeletePreconditionFailed):
				flash = T(i18n.TextDeletePrecondition)
			}
			cookie.Path = path.String()
			cookie.MaxAge = -1
			cookie.Expires = time.Unix(0, 0)
			http.SetCookie(w, cookie)
			return
		}
	}
	return
}
