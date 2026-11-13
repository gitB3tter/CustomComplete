package transport

import (
	"customcomplete/internal/i18n"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/weblinking"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"time"

	"golang.org/x/text/language"
)

func newCacheContent(b []byte, typ mediatypes.Type) PrebakedContent {
	return PrebakedContent{
		body:          b,
		etag:          GetEtagStrongBytes(b),
		contentLength: len(b),
		contentType:   typ,
	}
}

type PrebakedContent struct {
	body          []byte
	etag          ETagStrong
	contentLength int
	contentType   mediatypes.Type
}

func NewCache(path weblinking.Path, escape map[language.Tag]weblinking.Link, links map[language.Tag]weblinking.Links) (cache PrebakedCache) {
	cache.path = path
	cache.escape = escape
	cache.links = links
	cache.content = make(map[mediatypes.Type]map[language.Tag]PrebakedContent)
	return
}

type PrebakedCache struct {
	content        map[mediatypes.Type]map[language.Tag]PrebakedContent
	links          map[language.Tag]weblinking.Links
	escape         map[language.Tag]weblinking.Link
	supportedTypes mediatypes.Types
	path           weblinking.Path
}

func (cache PrebakedCache) With(typ mediatypes.Type, output map[language.Tag][]byte) PrebakedCache {
	cache = cache.add(output, typ)
	return cache
}

func (cache PrebakedCache) WithLinkset(links map[language.Tag]weblinking.Links) (PrebakedCache, error) {

	cache.content[mediatypes.Linkset] = make(map[language.Tag]PrebakedContent)
	cache.content[mediatypes.LinksetJSON] = make(map[language.Tag]PrebakedContent)
	for _, langTag := range i18n.SupportedLanguages {
		b := []byte(links[langTag].Linkset())
		etag, err := GetEtagStrong(b)
		if err != nil {
			return PrebakedCache{}, err
		}
		cache.content[mediatypes.Linkset][langTag] = PrebakedContent{
			body:          b,
			etag:          etag,
			contentLength: len(b),
			contentType:   mediatypes.Linkset,
		}
		b, err = links[langTag].MarshalJSON()
		if err != nil {
			return PrebakedCache{}, err
		}
		etag, err = GetEtagStrong(b)
		if err != nil {
			return PrebakedCache{}, err
		}
		cache.content[mediatypes.LinksetJSON][langTag] = PrebakedContent{
			body:          b,
			etag:          etag,
			contentLength: len(b),
			contentType:   mediatypes.LinksetJSON,
		}
	}
	cache.supportedTypes = append(cache.supportedTypes, mediatypes.Linkset, mediatypes.LinksetJSON)
	return cache, nil
}

func (cache PrebakedCache) GetSize(typ mediatypes.Type, langTag language.Tag) (int, bool) {
	_, ok := cache.content[typ]
	if !ok {
		return 0, false
	}
	content, ok := cache.content[typ][langTag]
	if !ok {
		return 0, false
	}
	return content.contentLength, true
}

func (cache PrebakedCache) RegisterOn(mux *http.ServeMux, pattern string) {
	if pattern == "" {
		pattern = cache.path.String()
	}
	mux.Handle("GET "+pattern, NoBody(cache.get))
	mux.Handle("OPTIONS "+pattern, NoBody(cache.options))
}

func (cache PrebakedCache) options(w http.ResponseWriter, r *http.Request) {
	langTag := i18n.LangTagFromContext(r.Context())
	w.Header().Set("Link", append(
		slices.Clone(cache.links[langTag]),
		weblinking.AllAlternates(cache.path, cache.supportedTypes...)...,
	).String())

	w.Header().Set("Allow", "GET, HEAD, OPTIONS")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Link, Allow, Vary")

	w.WriteHeader(http.StatusNoContent)
}

func (cache PrebakedCache) get(w http.ResponseWriter, r *http.Request) {
	langTag := i18n.LangTagFromContext(r.Context())
	responseType, proceed := NegotiateOr406(w, r, cache.escape[langTag], slices.Collect(maps.Keys(cache.content)))
	if !proceed {
		return
	}
	content, ok := cache.content[responseType][langTag]
	if !ok {
		InternalError(w, r, fmt.Errorf("No content for type %s lang %s", responseType.String(), langTag.String()), cache.escape[langTag])
		return
	}
	SetCachePublic(w.Header(), time.Hour, time.Hour*24)
	w.Header().Set("Link", append(
		slices.Clone(cache.links[langTag]),
		weblinking.Alternates(cache.path, langTag, responseType, cache.supportedTypes...)...,
	).String())
	content.ServeHTTP(w, r)
}

func (p PrebakedContent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ETag", string(p.etag))
	if EtagMatchStrong(r.Header.Values("If-None-Match"), p.etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	Write(r.Context(), w, p.body, p.contentType, http.StatusOK)
}

func (cache PrebakedCache) add(bakedGood map[language.Tag][]byte, typ mediatypes.Type) PrebakedCache {
	cache.content[typ] = make(map[language.Tag]PrebakedContent)
	for lang, b := range bakedGood {
		cache.content[typ][lang] = newCacheContent(b, typ)
	}
	cache.supportedTypes = append(cache.supportedTypes, typ)
	return cache
}
