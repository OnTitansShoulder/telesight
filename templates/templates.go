package templates

import (
	"fmt"
	"html/template"
)

const (
	StreamTemplate = "stream.gtpl"
)

func ParseAllTemplates(basePath string) *template.Template {
	templateList := []string{
		StreamTemplate,
	}
	// resolve the correct file path
	for i := 0; i < len(templateList); i++ {
		templateList[i] = fmt.Sprintf("%s/templates/%s", basePath, templateList[i])
	}
	return template.Must(template.ParseFiles(templateList...))
}
