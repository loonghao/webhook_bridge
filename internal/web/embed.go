package web

import (
	"html/template"
	"net/http"

	webpkg "github.com/loonghao/webhook_bridge/web-nextjs"
)

// Path constants for web resources
const (
	AssetsPath  = "/assets"
	FaviconPath = "/favicon.ico"
	IndexPath   = "index.html"
)

// EmbeddedAssets interface for accessing embedded web resources
type EmbeddedAssets interface {
	GetIndexHTML() string
	GetFaviconData() []byte
	GetJSFile() []byte
	GetCSSFile() []byte
}

// embeddedAssets holds the embedded resources
type embeddedAssets struct {
	indexHTML   string
	faviconData []byte
	jsFile      []byte
	cssFile     []byte
}

// NewEmbeddedAssets creates a new embedded assets instance
func NewEmbeddedAssets(indexHTML string, faviconData []byte, jsFile []byte, cssFile []byte) EmbeddedAssets {
	return &embeddedAssets{
		indexHTML:   indexHTML,
		faviconData: faviconData,
		jsFile:      jsFile,
		cssFile:     cssFile,
	}
}

func (e *embeddedAssets) GetIndexHTML() string {
	return e.indexHTML
}

func (e *embeddedAssets) GetFaviconData() []byte {
	return e.faviconData
}

func (e *embeddedAssets) GetJSFile() []byte {
	return e.jsFile
}

func (e *embeddedAssets) GetCSSFile() []byte {
	return e.cssFile
}

// Global embedded assets instance
var Assets EmbeddedAssets

// GetWebFS returns a simple file server for embedded assets
func GetWebFS() http.FileSystem {
	// For now, return nil - we'll handle static files directly in handlers
	return nil
}

// GetIndexTemplate returns the index.html template from embedded resources
func GetIndexTemplate() (*template.Template, error) {
	// Try to use embedded assets first
	if Assets != nil {
		tmpl, err := template.New("dashboard").Parse(Assets.GetIndexHTML())
		if err != nil {
			return nil, err
		}
		return tmpl, nil
	}

	// Fallback to direct embedded resources from Next.js package
	indexHTML, err := webpkg.GetIndexHTML()
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("dashboard").Parse(indexHTML)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// GetFaviconData returns the favicon.ico data from embedded resources
func GetFaviconData() ([]byte, error) {
	if Assets != nil {
		return Assets.GetFaviconData(), nil
	}
	// Try to get favicon from Next.js package
	return webpkg.GetFaviconData()
}

// GetJSFile returns the embedded JavaScript file
func GetJSFile() []byte {
	if Assets != nil {
		return Assets.GetJSFile()
	}
	// Fallback to Next.js main JS file
	jsData, err := webpkg.GetMainJS()
	if err != nil {
		return []byte{}
	}
	return jsData
}

// GetCSSFile returns the embedded CSS file
func GetCSSFile() []byte {
	if Assets != nil {
		return Assets.GetCSSFile()
	}
	// Fallback to Next.js main CSS file
	cssData, err := webpkg.GetMainCSS()
	if err != nil {
		return []byte{}
	}
	return cssData
}
