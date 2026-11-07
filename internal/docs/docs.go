package docs

import (
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
	"net/url"
)

type attachment struct {
	Path        *url.URL
	Filename    string
	FileSizeFmt string
	Type        mediatypes.Type
}

type collection struct {
	Base  ui.BaseDataHTML
	Items weblinking.Links
}

func GetPath(p path) *url.URL {
	return &url.URL{
		Path: string(p),
	}
}

func docsPath() *url.URL {
	return (&url.URL{Path: "/"}).JoinPath("docs")
}
