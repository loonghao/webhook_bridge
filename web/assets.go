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

//go:embed dist/assets/index.BoJitokV.js
var JSFile []byte

//go:embed dist/assets/index.MysGidYy.css
var CSSFile []byte
