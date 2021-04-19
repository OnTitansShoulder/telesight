package templates

import (
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func ParseAllTemplates(basePath string) *template.Template {
	templateDirPath := filepath.Join(basePath, "templates")
	files, err := ioutil.ReadDir(templateDirPath)
	if err != nil {
		log.Printf("Failed to read files from templateDir=%s: %v", templateDirPath, err)
	}
	// resolve the correct file path
	var templateFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".gtpl") {
			templateFiles = append(templateFiles, filepath.Join(templateDirPath, file.Name()))
		}
	}
	return template.Must(template.ParseFiles(templateFiles...))
}
