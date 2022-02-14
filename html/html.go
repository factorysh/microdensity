package html

import (
	"bytes"
	"embed"
	"html/template"
	_template "html/template"
	"net/http"
)

var (
	// used to ensure embed import
	_ embed.FS
	//go:embed templates
	templatesFS embed.FS
)

// Page wraps a page content to be used by the application
type Page struct {
	ServicesList []string
	Detail       string
	Domain       string
	Partial      Partial
	Inner        _template.HTML
}

// Render renders an application page using templates
func (p *Page) Render(w http.ResponseWriter) error {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	tmpl, err := template.ParseFS(templatesFS, "templates/*")
	if err != nil {
		return err
	}

	p.Inner, err = p.Partial.Render()
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, "application.html", p)
}

// Partial page to add to the application page
type Partial struct {
	Data     interface{}
	Template string
}

// Render a partial web page into a usable template.HTML
func (c *Partial) Render() (_template.HTML, error) {
	tmpl, err := _template.New("inner").Parse(c.Template)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, c.Data)
	if err != nil {
		return "", err
	}

	return template.HTML(buffer.Bytes()), err
}
