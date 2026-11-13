package weblinking

import (
	"customcomplete/internal/httpsyntax"
	"customcomplete/internal/mediatypes"
	"encoding/json"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/text/language"
)

// IANA registered Link relation or extension rel
type linkrel string

const (
	About          linkrel = "about"
	Alternate      linkrel = "alternate"
	Collection     linkrel = "collection"
	CreateForm     linkrel = "create-form"
	DescribedBy    linkrel = "describedby"
	Describes      linkrel = "describes"
	EditForm       linkrel = "edit-form"
	EditMedia      linkrel = "edit-media"
	Enclosure      linkrel = "enclosure"
	Item           linkrel = "item"
	Linkset        linkrel = "linkset"
	Next           linkrel = "next"
	Search         linkrel = "search"
	Self           linkrel = "self"
	ServiceDesc    linkrel = "service-desc"
	ServiceDoc     linkrel = "service-doc"
	Status         linkrel = "status"
	Up             linkrel = "up"
	deleteFormPath         = "/docs/rels/delete-form"
)

var (
	DeleteForm linkrel = linkrel((&url.URL{Path: "/"}).JoinPath("docs", "rels", "delete-form").String())
)

// value in Link header and linkset
type Link struct {
	rel       linkrel         `json:"-"`
	target    url.URL         `json:"-"`
	Context   *url.URL        `json:"-"`
	Type      mediatypes.Type `json:"type,omitzero"`
	Language  language.Tag    `json:"-"`
	Title     string          `json:"title,omitempty"`
	titleLang language.Tag    `json:"-"`
}

type titleExtended struct {
	Value    string `json:"value"`
	Language string `json:"language"`
}

type linkTargetObject struct {
	Target        url.URL         `json:"-"`
	Type          mediatypes.Type `json:"type,omitzero"`
	Language      []string        `json:"hreflang,omitempty"`
	Title         string          `json:"title,omitempty"`
	TitleExtended []titleExtended `json:"title*,omitempty"`
}

func Init(origin *url.URL) {
	if origin == nil {
		panic("weblinking.Init: nil origin")
	}

	DeleteForm = linkrel(origin.JoinPath(deleteFormPath).String())
}

func New(rel linkrel, target Path) Link {
	return Link{
		rel:    rel,
		target: target.URL(),
	}
}

func Breadcrumb(target Path, title string) Link {
	return Link{
		target: target.URL(),
		Title:  title,
	}
}

func (l Link) Rel() linkrel {
	return l.rel
}

func (l Link) Target() string {
	return l.target.String()
}

func (l Link) Lang() string {
	if l.Language == language.Und {
		return ""
	} else {
		return l.Language.String()
	}
}

func (l Link) WithContext(context *url.URL) Link {
	l.Context = context
	return l
}
func (l Link) WithType(typ mediatypes.Type) Link {
	l.Type = typ
	return l
}
func (l Link) WithLang(lang language.Tag) Link {
	l.Language = lang
	return l
}
func (l Link) WithTitle(title string) Link {
	l.Title = title
	return l
}
func (l Link) WithTitleLang(lang language.Tag) Link {
	l.titleLang = lang
	return l
}

func (l Link) WithQueryVal(key, val string) Link {
	u := l.target
	q := u.Query()
	q.Add(key, val)
	u.RawQuery = q.Encode()
	l.target = u
	return l
}

func (l Link) WithQuery(q url.Values) Link {
	u := l.target
	u.RawQuery = q.Encode()
	l.target = u
	return l
}

func (l Link) String() string {
	return l.stringWithSeparator("; ")
}

func (l Link) IsZero() bool {
	return l.target == (url.URL{})
}

func (l Link) MarshalBinary() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l Link) stringWithSeparator(sep string) string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(l.target.String())
	b.WriteByte('>')
	if l.Context != nil {
		if context := l.Context.String(); context != "" {
			b.WriteString(sep)
			b.WriteString(`anchor="`)
			b.WriteString(context)
			b.WriteByte('"')
		}
	}
	b.WriteString(sep)
	b.WriteString(`rel="`)
	b.WriteString(string(l.rel))
	b.WriteByte('"')
	if !l.Type.IsZero() {
		b.WriteString(sep)
		b.WriteString(`type="`)
		b.WriteString(l.Type.String())
		b.WriteByte('"')
	}
	if l.Language != language.Und {
		b.WriteString(sep)
		b.WriteString(`hreflang="`)
		b.WriteString(l.Language.String())
		b.WriteByte('"')
	}
	if l.Title != "" {
		b.WriteString(sep)
		if httpsyntax.IsASCII(l.Title) && l.titleLang == language.Und {
			b.WriteString(`title="`)
			b.WriteString(httpsyntax.QuoteEscape(l.Title))
			b.WriteByte('"')
		} else {
			b.WriteString(`title*=UTF-8'`)
			if l.titleLang != language.Und {
				b.WriteString(l.titleLang.String())
			}
			b.WriteByte('\'')
			b.WriteString(httpsyntax.PercentEncode(l.Title))
		}
	}

	return b.String()
}

func (l Link) linkTargetObject() (t linkTargetObject) {
	t.Target = l.target
	t.Type = l.Type

	if l.Language != language.Und {
		t.Language = append(t.Language, l.Language.String())
	}
	if l.titleLang != language.Und {
		t.TitleExtended = append(t.TitleExtended, titleExtended{
			Value:    l.Title,
			Language: l.titleLang.String(),
		})
	} else {
		t.Title = l.Title
	}

	return
}

// slice of Links that can be output both in Link Header aswell as as linkset body
type Links []Link

func (links Links) String() string {
	var b strings.Builder
	for i, link := range links {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(link.String())
	}
	return b.String()
}

// stops before exceeding maxBytes
func (links Links) StringUpTo(maxBytes int) string {
	var b strings.Builder
	if b.Cap() < maxBytes {
		b.Grow(maxBytes - b.Cap() + 1)
	}
	for i, link := range links {
		s := link.String()
		sepLen := 0
		if i > 0 {
			sepLen = 2 // ", "
		}
		if b.Len()+sepLen+len(s) > maxBytes {
			break
		}
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(s)
	}
	return b.String()
}

func (links Links) MarshalJSON() ([]byte, error) {
	return json.Marshal(links.linksetObject())
}

func (links Links) Linkset() string {
	var b strings.Builder
	for i, link := range links {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString(link.stringWithSeparator(";\n"))
	}
	return b.String()
}

func (links Links) linksetObject() linksetObject {

	//link.Context.String()
	type contextKey string

	contextObjects := make(map[contextKey]*linkContextObject, len(links))
	for _, link := range links {
		var ctxtKey contextKey
		if link.Context != nil {
			ctxtKey = contextKey(link.Context.String())
		}

		ctxObj := contextObjects[ctxtKey]
		if ctxObj == nil {
			ctxObj = &linkContextObject{
				Context: link.Context,
				Rels:    make(map[linkrel][]linkTargetObject),
			}
			contextObjects[ctxtKey] = ctxObj
		}

		ctxObj.Rels[link.rel] = appendOrMerge(ctxObj.Rels[link.rel], link)
	}

	contexts := make([]contextKey, 0, len(contextObjects))
	for key := range contextObjects {
		contexts = append(contexts, key)
	}
	slices.Sort(contexts)

	linkset := make([]linkContextObject, 0, len(contexts))
	for _, context := range contexts {
		linkset = append(linkset, *contextObjects[context])
	}

	return linksetObject{
		Linkset: linkset,
	}

}

type linksetObject struct {
	Linkset []linkContextObject `json:"linkset"`
}

type linkContextObject struct {
	Context *url.URL
	Rels    map[linkrel][]linkTargetObject
}

func appendOrMerge(targets []linkTargetObject, l Link) []linkTargetObject {
	obj := l.linkTargetObject()
	href := l.target.String()
	for i := range targets {
		target := &targets[i]
		if target.Target.String() == href && target.Type == obj.Type && target.Title == obj.Title {
			target.Language = append(target.Language, obj.Language...)
			target.TitleExtended = append(target.TitleExtended, obj.TitleExtended...)
			return targets
		}
	}
	return append(targets, l.linkTargetObject())
}

func (c linkContextObject) MarshalJSON() ([]byte, error) {
	out := make(map[string]any, len(c.Rels)+1)
	if c.Context != nil {
		if anchor := c.Context.String(); anchor != "" {
			out["anchor"] = anchor
		}
	}
	for rel, targets := range c.Rels {
		out[string(rel)] = targets
	}
	return json.Marshal(out)
}

func (l linkTargetObject) MarshalJSON() ([]byte, error) {
	type Alias linkTargetObject
	return json.Marshal(struct {
		Href string `json:"href"`
		Alias
	}{
		Href:  l.Target.String(),
		Alias: Alias(l),
	})
}
