package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
)

func (a *app) handleBlogPost(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/_posts/")
	if r.URL.Path == "/" {
		a.handleIndex(w, r)
		return
	}

	a.infoLog.Printf("Resource requested: %s\n", path)

	mapMutex.RLock()
	filePath, ok := urlToFileMap[path]
	mapMutex.RUnlock()

	w.Header().Set("Permission-Policy", "unload()")

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
		"Title":      string(title),
		"Content":    template.HTML(output),
		"Name":       a.config["name"],
		"Email":      a.config["email"],
		"Time":       time.Now().UTC().Format("15:04"),
		"StatusText": a.config["status"],

		// todo: add date and author to template
		// add some front matter to the markdown files
	})
	if err != nil {
		a.errorLog.Printf("Template error: %v\n", err)
	}
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	posts := []blogPost{}

	mapMutex.RLock()
	for url, filePath := range urlToFileMap {
		title := strings.Replace(strings.TrimPrefix(filePath, "_posts/"), "-", " ", -1)
		// though golang.org/x/text/case can be used, it's not worth the dependency
		title = strings.Title(strings.TrimSuffix(title, ".md"))
		posts = append(posts, blogPost{
			Title: title,
			URL:   url,
		})
	}
	mapMutex.RUnlock()

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Title < posts[j].Title
	})

	err := indexTmpl.ExecuteTemplate(w, "index", map[string]interface{}{
		"Posts":      posts,
		"Name":       a.config["name"],
		"Email":      a.config["email"],
		"Time":       time.Now().UTC().Format("15:04"),
		"StatusText": a.config["status"],
	})
	if err != nil {
		a.errorLog.Printf("Template execution error: %v\n", err)
	}
}
