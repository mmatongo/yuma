package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	urlToFileMap  = make(map[string]string)
	templateFiles = []string{
		"./ui/templates/base.layout.gohtml",
		"./ui/templates/header.partial.gohtml",
		"./ui/templates/nav.partial.gohtml",
		"./ui/templates/footer.partial.gohtml",
	}

	indexFiles = []string{
		"./ui/templates/index.page.gohtml",
		"./ui/templates/header.partial.gohtml",
		"./ui/templates/nav.partial.gohtml",
		"./ui/templates/footer.partial.gohtml",
	}
	tmpl      = template.Must(template.ParseFiles(templateFiles...))
	indexTmpl = template.Must(template.ParseFiles(indexFiles...))
	mapMutex  = &sync.RWMutex{}
)

type app struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	watcher  *fsnotify.Watcher
	config   map[string]interface{}
}

type blogPost struct {
	Title string
	URL   string
}

func main() {
	a := &app{
		errorLog: log.New(os.Stderr, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile),
		infoLog:  log.New(os.Stdout, "[INFO]\t", log.Ldate|log.Ltime),
		config:   make(map[string]interface{}),
	}

	a.initURLToFileMap()
	a.watchFiles()

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./ui/static"))))
	mux.HandleFunc("/_posts/", a.handleBlogPost)
	mux.HandleFunc("/", a.handleIndex)

	a.initConfig()
	if a.config != nil {
		a.infoLog.Printf("Config file was successfully loaded\n")
	}
	port := os.Getenv("PORT")

	srv := &http.Server{
		Addr:     ":" + port,
		Handler:  mux,
		ErrorLog: a.errorLog,
	}

	defer func() {
		if a.watcher != nil {
			a.watcher.Close()
		}
	}()

	a.infoLog.Printf("Server starting at port %s", port)
	err := srv.ListenAndServe()
	if err != nil {
		a.errorLog.Fatal("ListenAndServe: ", err)
	}
}
