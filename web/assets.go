package web

import (
	_ "embed"
)

// Embed specific files from web/dist directory
//
//go:embed dist/index.html
var IndexHTML string

//go:embed dist/favicon.ico
var FaviconData []byte

//go:embed dist/assets/index.CA1_AtTQ.js
var JSFile []byte

//go:embed dist/assets/index.tn0RQdqM.css
var CSSFile []byte
