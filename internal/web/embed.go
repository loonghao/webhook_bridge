package web

import (
	"html/template"
	"net/http"

	webpkg "github.com/loonghao/webhook_bridge/web"
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

	// Fallback to direct embedded resources from web package
	// Use the embedded assets directly from the web package
	tmpl, err := template.New("dashboard").Parse(webpkg.IndexHTML)
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
	// Fallback to direct embedded resources from web package
	return webpkg.FaviconData, nil
}

// GetJSFile returns the embedded JavaScript file
func GetJSFile() []byte {
	if Assets != nil {
		return Assets.GetJSFile()
	}
	// Fallback to direct embedded resources from web package
	return webpkg.JSFile
}

// GetCSSFile returns the embedded CSS file
func GetCSSFile() []byte {
	if Assets != nil {
		return Assets.GetCSSFile()
	}
	// Fallback to direct embedded resources from web package
	return webpkg.CSSFile
}
