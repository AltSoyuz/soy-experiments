package web

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed layouts/*.html components/*.html pages/*.html
var htmlFiles embed.FS

const defaultLayout = "base.html"

// TemplateCache provides a more efficient template management system
type TemplateCache struct {
	mu       sync.RWMutex
	cache    map[string]*template.Template
	funcMap  template.FuncMap
	htmlBase *template.Template
}

// NewTemplateCache creates an optimized template cache
func NewTemplateCache(additionalFuncs ...template.FuncMap) *TemplateCache {
	// Merge additional func maps if provided
	mergedFuncMap := template.FuncMap{
		"safe":    func(s string) template.HTML { return template.HTML(s) },
		"toUpper": strings.ToUpper,
		"toLower": strings.ToLower,
		"trimStr": strings.TrimSpace,
		"now":     time.Now,
		"upperFirst": func(s string) string {
			if s == "" {
				return ""
			} else {
				return strings.ToUpper(s[:1]) + s[1:]
			}
		},
		"formatDate": func(t time.Time, format string) string {
			return t.Format(format)
		},
	}

	for _, fm := range additionalFuncs {
		for k, v := range fm {
			mergedFuncMap[k] = v
		}
	}

	return &TemplateCache{
		cache:   make(map[string]*template.Template),
		funcMap: mergedFuncMap,
	}
}

// PrecompileTemplates loads and caches all templates upfront
func (tc *TemplateCache) PrecompileTemplates() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Precompile base layout with global functions
	baseLayoutTmpl, err := template.New("base").Funcs(tc.funcMap).ParseFS(htmlFiles, "layouts/"+defaultLayout)
	if err != nil {
		return fmt.Errorf("failed to parse base layout: %w", err)
	}
	tc.htmlBase = baseLayoutTmpl

	// Add base layout to cache
	tc.cache["base"] = baseLayoutTmpl

	// Read all components
	components, err := htmlFiles.ReadDir("components")
	if err != nil {
		return fmt.Errorf("failed to read components directory: %w", err)
	}

	// Parse and cache components
	for _, component := range components {
		if !component.IsDir() && strings.HasSuffix(component.Name(), ".html") {
			componentPath := filepath.Join("components", component.Name())
			if err := tc.parseAndCacheTemplate(componentPath, true); err != nil {
				return err
			}
		}
	}

	// Read all pages
	pages, err := htmlFiles.ReadDir("pages")
	if err != nil {
		return fmt.Errorf("failed to read pages directory: %w", err)
	}

	// Parse and cache pages
	for _, page := range pages {
		if !page.IsDir() && strings.HasSuffix(page.Name(), ".html") {
			pagePath := filepath.Join("pages", page.Name())
			if err := tc.parseAndCacheTemplate(pagePath, false); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseAndCacheTemplate efficiently parses and caches a template
func (tc *TemplateCache) parseAndCacheTemplate(path string, isComponent bool) error {
	name := strings.TrimSuffix(filepath.Base(path), ".html")

	// Determine which files to parse based on template type
	var files []string
	if isComponent {
		files = []string{"components/*.html", path}
	} else {
		files = []string{"layouts/" + defaultLayout, "components/*.html", path}
	}

	// Parse template with global functions
	tmpl, err := template.New(name).Funcs(tc.funcMap).ParseFS(htmlFiles, files...)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	tc.cache[name] = tmpl
	return nil
}

// GetTemplate retrieves a template with thread-safe read access
func (tc *TemplateCache) GetTemplate(name string) (*template.Template, error) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	tmpl, exists := tc.cache[name]
	if !exists {
		return nil, fmt.Errorf("template %s not found", name)
	}
	return tmpl, nil
}

// RenderTemplate provides a more robust template rendering
func (tc *TemplateCache) RenderTemplate(w io.Writer, name, block string, data any) {
	tmpl, err := tc.GetTemplate(name)
	if err != nil {
		fmt.Printf("error getting template: %v\n", err)
		return
	}

	// Improved error handling with context
	if err := tmpl.ExecuteTemplate(w, block, data); err != nil {
		fmt.Printf("error executing template: %v\n", err)
		return
	}
}

// Singleton cache for global access
var TemplateSystem = NewTemplateCache()

// RenderPage renders a full page with the base layout
func RenderPage(w io.Writer, name string, data any) {
	TemplateSystem.RenderTemplate(w, name, defaultLayout, data)
}

// RenderComponent renders a reusable component
func RenderComponent(w io.Writer, name, block string, data any) {
	TemplateSystem.RenderTemplate(w, name, block, data)
}
