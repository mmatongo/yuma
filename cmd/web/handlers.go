package main

import (
	"html/template"
	"net/http"
	"os"
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
	err = tmpl.Execute(w, map[string]interface{}{
		"Title":   "Your Blog Title Here",
		"Content": template.HTML(output),
	})
	if err != nil {
		a.errorLog.Printf("Template error: %v\n", err)
	}
}
