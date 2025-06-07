package web

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"time"
)

// Embed the entire Next.js dist directory and public assets
//
//go:embed dist public
var NextJSAssets embed.FS

// GetNextJSFS returns the embedded Next.js filesystem
func GetNextJSFS() fs.FS {
	// Return the dist subdirectory
	distFS, err := fs.Sub(NextJSAssets, "dist")
	if err != nil {
		// Fallback to the full filesystem if subdirectory fails
		return NextJSAssets
	}
	return distFS
}

// DebugListFiles lists all files in the embedded filesystem for debugging
func DebugListFiles() {
	fmt.Println("=== Debugging embedded filesystem ===")

	// List files in the root
	fmt.Println("Files in root:")
	if entries, err := NextJSAssets.ReadDir("."); err == nil {
		for _, entry := range entries {
			fmt.Printf("  %s (dir: %v)\n", entry.Name(), entry.IsDir())
		}
	} else {
		fmt.Printf("Error reading root: %v\n", err)
	}

	// List files in dist
	fmt.Println("\nFiles in dist:")
	if entries, err := NextJSAssets.ReadDir("dist"); err == nil {
		for _, entry := range entries {
			fmt.Printf("  %s (dir: %v)\n", entry.Name(), entry.IsDir())
		}
	} else {
		fmt.Printf("Error reading dist: %v\n", err)
	}

	// List files in dist/next
	fmt.Println("\nFiles in dist/next:")
	if entries, err := NextJSAssets.ReadDir("dist/next"); err == nil {
		for _, entry := range entries {
			fmt.Printf("  %s (dir: %v)\n", entry.Name(), entry.IsDir())
		}
	} else {
		fmt.Printf("Error reading dist/next: %v\n", err)
	}

	// List files in dist/next/static
	fmt.Println("\nFiles in dist/next/static:")
	if entries, err := NextJSAssets.ReadDir("dist/next/static"); err == nil {
		for _, entry := range entries {
			fmt.Printf("  %s (dir: %v)\n", entry.Name(), entry.IsDir())
		}
	} else {
		fmt.Printf("Error reading dist/next/static: %v\n", err)
	}

	// List files in dist/next/static/css
	fmt.Println("\nFiles in dist/next/static/css:")
	if entries, err := NextJSAssets.ReadDir("dist/next/static/css"); err == nil {
		for _, entry := range entries {
			fmt.Printf("  %s (dir: %v)\n", entry.Name(), entry.IsDir())
		}
	} else {
		fmt.Printf("Error reading dist/next/static/css: %v\n", err)
	}

	// List files in dist/next/static/chunks
	fmt.Println("\nFiles in dist/next/static/chunks:")
	if entries, err := NextJSAssets.ReadDir("dist/next/static/chunks"); err == nil {
		for _, entry := range entries {
			fmt.Printf("  %s (dir: %v)\n", entry.Name(), entry.IsDir())
		}
	} else {
		fmt.Printf("Error reading dist/next/static/chunks: %v\n", err)
	}

	fmt.Println("=== End debugging ===")
}

// GetIndexHTML returns the index.html content from Next.js build
func GetIndexHTML() (string, error) {
	data, err := NextJSAssets.ReadFile("dist/index.html")
	if err != nil {
		return "", err
	}

	// Convert Next.js static export HTML to work with our server
	html := string(data)

	// Path replacement is handled by build-and-fix.js during build process
	// HTML files already contain correct /next/static/ paths
	return html, nil
}

// GetMainCSS returns the main CSS file content
func GetMainCSS() ([]byte, error) {
	// Next.js generates CSS files with hashes, so we need to find the CSS file
	entries, err := NextJSAssets.ReadDir("dist/next/static/css")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[len(entry.Name())-4:] == ".css" {
			return NextJSAssets.ReadFile("dist/next/static/css/" + entry.Name())
		}
	}

	return nil, fs.ErrNotExist
}

// GetMainJS returns the main JavaScript file content
func GetMainJS() ([]byte, error) {
	// Next.js generates multiple JS files, we'll return the main app JS
	entries, err := NextJSAssets.ReadDir("dist/next/static/chunks")
	if err != nil {
		return nil, err
	}

	// Look for the main app JS file
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 3 && entry.Name()[len(entry.Name())-3:] == ".js" {
			// Return the first JS file found (could be improved to find specific main file)
			return NextJSAssets.ReadFile("dist/next/static/chunks/" + entry.Name())
		}
	}

	return nil, fs.ErrNotExist
}

// GetFaviconData returns the favicon.ico content
func GetFaviconData() ([]byte, error) {
	// Try to read favicon from public directory first
	if data, err := NextJSAssets.ReadFile("public/favicon.ico"); err == nil {
		return data, nil
	}

	// Try to read from dist directory
	if data, err := NextJSAssets.ReadFile("dist/favicon.ico"); err == nil {
		return data, nil
	}

	// Return empty data if no favicon found
	return []byte{}, fs.ErrNotExist
}

// GetCSSLoadingStatus returns detailed CSS loading status for debugging
func GetCSSLoadingStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// Check CSS directory accessibility
	cssDir := "dist/next/static/css"
	if entries, err := NextJSAssets.ReadDir(cssDir); err == nil {
		status["css_directory_accessible"] = true
		status["css_file_count"] = len(entries)

		// List all CSS files
		cssFiles := make([]map[string]interface{}, 0)
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".css") {
				fileInfo := map[string]interface{}{
					"name": entry.Name(),
					"path": cssDir + "/" + entry.Name(),
				}

				// Try to get file size
				if info, err := entry.Info(); err == nil {
					fileInfo["size"] = info.Size()
					fileInfo["modified"] = info.ModTime().Format("2006-01-02 15:04:05")
				}

				// Try to read file content
				if data, err := NextJSAssets.ReadFile(cssDir + "/" + entry.Name()); err == nil {
					fileInfo["readable"] = true
					fileInfo["content_size"] = len(data)
					// Preview first 200 characters
					content := string(data)
					if len(content) > 200 {
						content = content[:200] + "..."
					}
					fileInfo["preview"] = content
				} else {
					fileInfo["readable"] = false
					fileInfo["read_error"] = err.Error()
				}

				cssFiles = append(cssFiles, fileInfo)
			}
		}
		status["css_files"] = cssFiles
	} else {
		status["css_directory_accessible"] = false
		status["css_directory_error"] = err.Error()
	}

	// Check legacy directory
	legacyDir := "dist/_next/static/css"
	if entries, err := NextJSAssets.ReadDir(legacyDir); err == nil {
		status["legacy_directory_exists"] = true
		status["legacy_file_count"] = len(entries)
	} else {
		status["legacy_directory_exists"] = false
		status["legacy_error"] = err.Error()
	}

	// Test GetMainCSS function
	if cssData, err := GetMainCSS(); err == nil {
		status["get_main_css_success"] = true
		status["main_css_size"] = len(cssData)
	} else {
		status["get_main_css_success"] = false
		status["get_main_css_error"] = err.Error()
	}

	// Test GetIndexHTML function and check for CSS references
	if htmlData, err := GetIndexHTML(); err == nil {
		status["get_index_html_success"] = true
		status["index_html_size"] = len(htmlData)

		// Find CSS references in HTML
		cssRefs := make([]string, 0)
		lines := strings.Split(htmlData, "\n")
		for _, line := range lines {
			if strings.Contains(line, ".css") && (strings.Contains(line, "href=") || strings.Contains(line, "link")) {
				cssRefs = append(cssRefs, strings.TrimSpace(line))
			}
		}
		status["css_references_in_html"] = cssRefs
		status["css_reference_count"] = len(cssRefs)
	} else {
		status["get_index_html_success"] = false
		status["get_index_html_error"] = err.Error()
	}

	return status
}

// LogCSSLoadingAttempt logs a CSS loading attempt with detailed information
func LogCSSLoadingAttempt(requestPath string, success bool, errorMsg string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	if success {
		fmt.Printf("[%s] CSS LOAD SUCCESS: %s\n", timestamp, requestPath)
	} else {
		fmt.Printf("[%s] CSS LOAD FAILED: %s - Error: %s\n", timestamp, requestPath, errorMsg)
	}
}
