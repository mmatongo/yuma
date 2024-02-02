package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
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
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					a.infoLog.Println("modified file:", event.Name)
					updateURLToFileMap(event.Name)
				case event.Op&fsnotify.Create == fsnotify.Create:
					a.infoLog.Println("created file:", event.Name)
					updateURLToFileMap(event.Name)

					if d, err := os.Stat(event.Name); err == nil && d.IsDir() {
						a.watchDirRecursively(event.Name)
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					a.infoLog.Println("removed file:", event.Name)
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

	a.watchDirRecursively("_posts")

	// TODO: watch for changes in the templates directory
	// TODO: make it so that created directories are watched and their contents are watched
}

// an abomination spawned from the depths of ai hell with some help from stackoverflow and the golang docs
// this is a solution to a nonexistent problem but i'm leaving it here because it's cool and i'm proud of it (famous last words)
// part of why this exists is because I intend for the binary to be used in a container or server that I will push files to
// and I don't want to have to restart the server every time I push a new file
func (a *app) watchDirRecursively(dirPath string) {
	// Watch the directory.
	err := a.watcher.Add(dirPath)
	if err != nil {
		a.errorLog.Println("error watching directory:", dirPath, "error:", err)
		return
	}

	// iterate through the directory and watch all subdirectories
	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			a.errorLog.Println("error walking through directory:", dirPath, "error:", err)
			return nil
		}

		// If it's a directory, watch it.
		if d.IsDir() {
			err := a.watcher.Add(path)
			if err != nil {
				a.errorLog.Println("error watching directory:", path, "error:", err)
			}
		}

		return nil
	})
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

func (a *app) initConfig() {
	config, err := os.ReadFile("config/config.yml")
	if err != nil {
		a.errorLog.Fatalf("Error reading config file: %v", err)
	}
	err = yaml.Unmarshal(config, &a.config)

	if err != nil {
		a.errorLog.Fatalf("Error unmarshalling config file: %v", err)
	}
}
