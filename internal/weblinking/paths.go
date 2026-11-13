package weblinking

import (
	"net/url"
)

type Path string

const (
	RootPath        Path = "/"
	DocsPath        Path = "/docs"
	DocsSpecsPath   Path = "/docs/specs"
	DocsSchemasPath Path = "/docs/schemas"
)

func (p Path) Join(elem ...string) Path {
	return Path(new(p.URL()).JoinPath(elem...).Path)
}

func (p Path) Link(rel linkrel) Link {
	return New(rel, p)
}

func (p Path) URL() url.URL {
	return url.URL{Path: string(p)}
}

func (p Path) String() string {
	return string(p)
}
