package ui

import (
	"embed"
)

//go:embed all:templates
var TemplatesFS embed.FS

//go:embed all:public
var PublicFS embed.FS
