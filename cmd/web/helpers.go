package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func (a *app) initURLToFileMap() {
	err := filepath.WalkDir("_posts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			updateURLToFileMap(path)
		}
		return nil
	})

	if err != nil {
		a.errorLog.Fatalf("Failed to walk _posts directory: %v", err)
	}
}

func (a *app) watchFiles() {
	var err error
	a.watcher, err = fsnotify.NewWatcher()

	if err != nil {
		a.errorLog.Fatal(err)
	}
	//defer a.watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-a.watcher.Events:
				if !ok {
					return
				}
				//a.infoLog.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
					a.infoLog.Println("modified file:", event.Name)
					updateURLToFileMap(event.Name)
				}
			case err, ok := <-a.watcher.Errors:
				if !ok {
					return
				}
				a.errorLog.Println("error:", err)
			}
		}
	}()

	err = a.watcher.Add("_posts")
	if err != nil {
		a.errorLog.Println(err.Error())
	}
}

func updateURLToFileMap(filePath string) {
	if filepath.Ext(filePath) != ".md" {
		return
	}

	urlPath := strings.TrimPrefix(filePath, "_posts/")
	urlPath = strings.TrimSuffix(urlPath, ".md")
	urlPath = strings.ReplaceAll(urlPath, "_", "-")

	mapMutex.Lock()
	defer mapMutex.Unlock()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		delete(urlToFileMap, urlPath)
	} else {
		urlToFileMap[urlPath] = filePath
	}
}
