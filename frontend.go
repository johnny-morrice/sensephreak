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

func (fr *frontend) script() string {
	return fr.getasset("data/script.js")
}

func (fr *frontend) srcmap() string {
        return fr.getasset("data/script.js.map")
}

func (fr *frontend) html() string {
	return fr.getasset("data/index.html")
}

func (fr *frontend) css() string {
	return fr.getasset("data/style.css")
}

func (fr *frontend) indexpage() []byte {
        if fr.cache != nil {
                return fr.cache
        }

	style := fr.css()
	html := fr.html()

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
        js := fr.script()

        w.Write([]byte(js))
}

func (fr *frontend) sourcemap(w http.ResponseWriter, r *http.Request) {
        m := fr.srcmap()

        w.Write([]byte(m))
}
