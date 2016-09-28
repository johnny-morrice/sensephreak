//go:generate gopherjs build -m -o data/script.js js/main.go
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
        cache []byte
}

func (fr *frontend) getasset(name string) string {
	template, err := Asset(name)

	if err != nil {
		panic(err)
	}

	return string(template)
}

func (fr *frontend) jqueryasset() string {
	return fr.getasset("data/jquery-3.1.1.min.js")
}

func (fr *frontend) scriptasset() string {
	return fr.getasset("data/script.js")
}

func (fr *frontend) srcmapasset() string {
        return fr.getasset("data/script.js.map")
}

func (fr *frontend) htmlasset() string {
	return fr.getasset("data/index.html")
}

func (fr *frontend) cssasset() string {
	return fr.getasset("data/style.css")
}

func (fr *frontend) indexpage() []byte {
        if fr.cache != nil {
                return fr.cache
        }

	style := fr.cssasset()
	html := fr.htmlasset()

	type Variables struct {
		Css        string
		Ports      []int
		Hostname   string
		Apiport    int
	}

	variables := Variables{
		Css:        style,
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

	fr.cache = buff.Bytes()

        return fr.cache
}

func (fr *frontend) index(w http.ResponseWriter, r *http.Request) {
	page := fr.indexpage()

	w.Write(page)
}

func (fr *frontend) javascript(w http.ResponseWriter, r *http.Request) {
        js := fr.scriptasset()

        w.Write([]byte(js))
}

func (fr *frontend) sourcemap(w http.ResponseWriter, r *http.Request) {
        m := fr.srcmapasset()

        w.Write([]byte(m))
}

func (fr *frontend) jquery(w http.ResponseWriter, r *http.Request) {
	jq := fr.jqueryasset()

	w.Write([]byte(jq))
}
