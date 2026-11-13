package docs

import (
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/ui"
	"customcomplete/internal/weblinking"
)

type attachment struct {
	Path        weblinking.Path
	Filename    string
	FileSizeFmt string
	Type        mediatypes.Type
}

type collection struct {
	Base  ui.BaseDataHTML
	Items weblinking.Links
}
