package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
)

func (a *app) handleBlogPost(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	mapMutex.RLock()
	filePath, ok := urlToFileMap[path]
	mapMutex.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		a.errorLog.Printf("Error reading file: %s, error: %v\n", filePath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	output := markdown.ToHTML(data, nil, nil)

	// should move this to a helper function
	dashTitle := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filepath.Base(filePath)))
	title := strings.ToUpper(strings.Replace(dashTitle, "-", " ", -1))

	err = tmpl.Execute(w, map[string]interface{}{
		"Title":   string(title),
		"Content": template.HTML(output),
		// todo: add date and author to template
		// add some front matter to the markdown files
	})
	if err != nil {
		a.errorLog.Printf("Template error: %v\n", err)
	}
}
