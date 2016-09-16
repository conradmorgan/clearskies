package view

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

type View struct {
	Title   string
	File    string
	Data    interface{}
	Headers []template.HTML
	Content template.HTML
}

func (v *View) AddHeader(header string, data ...interface{}) {
	if len(data) > 1 {
		log.Fatal("Add header supports at most one piece of data!")
	}
	tmpl, _ := template.New("").Parse(header)
	var input interface{}
	if len(data) == 1 {
		input = data[0]
	}
	var buf bytes.Buffer
	tmpl.Execute(&buf, input)
	v.Headers = append(v.Headers, template.HTML(buf.Bytes()))
}

func (v *View) Render(w http.ResponseWriter) {
	content, err := template.ParseFiles("templates/" + v.File)
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	err = content.Execute(&buf, v.Data)
	if err != nil {
		log.Fatal(err)
	}
	v.Content = template.HTML(buf.Bytes())
	base, _ := template.ParseFiles("templates/base.html")
	w.Header().Set("content-Type", "text/html")
	base.Execute(w, v)
}
