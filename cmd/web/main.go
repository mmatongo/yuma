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
	urlToFileMap = make(map[string]string)
	tmpl         = template.Must(template.ParseFiles("ui/templates/base.layout.gohtml"))
	mapMutex     = &sync.RWMutex{}
)

type app struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	watcher  *fsnotify.Watcher
}

func main() {
	a := &app{
		errorLog: log.New(os.Stderr, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile),
		infoLog:  log.New(os.Stdout, "[INFO]\t", log.Ldate|log.Ltime),
	}

	a.initURLToFileMap()
	a.watchFiles()

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./ui/static"))))
	mux.HandleFunc("/", a.handleBlogPost)

	srv := &http.Server{
		Addr:     ":8080",
		Handler:  mux,
		ErrorLog: a.errorLog,
	}

	defer func() {
		if a.watcher != nil {
			a.watcher.Close()
		}
	}()

	a.infoLog.Println("Server starting at :8080")
	err := srv.ListenAndServe()
	if err != nil {
		a.errorLog.Fatal("ListenAndServe: ", err)
	}
}