//go:generate gopherjs build -m -o data/script.js js/main.go
//go:generate rm data/script.js.map
//go:generate go-bindata data/
package main

import (
	"bytes"
	// We only generate from trusted data so text/template is fine.
	"net/http"
	"text/template"
)

type frontend struct {
	ports   []int
	host    string
	apiport int
}

func (fr *frontend) getasset(name string) string {
	template, err := Asset(name)

	if err != nil {
		panic(err)
	}

	return string(template)
}

func (fr *frontend) javascript() string {
	return fr.getasset("data/script.js")
}

func (fr *frontend) html() string {
	return fr.getasset("data/index.html")
}

func (fr *frontend) css() string {
	return fr.getasset("data/style.css")
}

func (fr *frontend) IndexPage() []byte {
	style := fr.css()
	js := fr.javascript()
	html := fr.html()

	type Variables struct {
		Css        string
		Javascript string
		Ports      []int
		Hostname   string
		Apiport    int
	}

	variables := Variables{
		Css:        style,
		Javascript: js,
		Ports:      fr.ports,
		Apiport:    fr.apiport,
		Hostname:   fr.host,
	}

	buff := &bytes.Buffer{}

	tmpl, err := template.New("index.html").Parse(html)

	if err != nil {
		panic(err)
	}

	tmpl.Execute(buff, variables)

	return buff.Bytes()
}

func (fr *frontend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page := fr.IndexPage()

	w.Write(page)
}
