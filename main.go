package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Folders   []string  `json:"folders"`
	ScaleStep int       `json:"scalestep"`
	MinWidth  int       `json:"minwidth"`
	Colors    [5]string `json:"colors"`
	Images    []string
	Port      int `json:"port"`
}

type TemplateData struct {
	Config Config
}

func main() {
	// Read and parse config
	config := readConfig("config.json")

	// Setup HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveGallery(w, config)
	})

	// Serve static files (images)
	for _, folder := range config.Folders {
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(folder))))
	}

	fmt.Println(fmt.Sprintf("starting on %d", config.Port))
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}

func readConfig(filename string) Config {
	configFile, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		os.Exit(1)
	}

	// Scan for images
	imageExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".bmp": true, ".webp": true, ".tiff": true,
	}

	for _, folder := range config.Folders {
		filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))
				if imageExtensions[ext] {
					fname := fmt.Sprintf("localhost:%d/static/%s", config.Port, filepath.Base(path))
					config.Images = append(config.Images, fname)
				}
			}
			return nil
		})
	}

	return config
}

func serveGallery(w http.ResponseWriter, config Config) {
	// Parse template
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare template data
	data := TemplateData{
		Config: config,
	}

	// Execute template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
