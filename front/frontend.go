package front
//go:generate go-bindata -pkg front data/

import (
        "bytes"
        // We only generate from trusted data so text/template is fine.
        "text/template"
        "net/http"
)

type Frontend struct {
        Ports []int
}

func (fr *Frontend) javascript() string {
        template, err := Asset("data/script.js")

        if err != nil {
                panic(err)
        }

        return string(template)
}

func (fr *Frontend) html() string {
        template, err := Asset("data/index.html")

        if err != nil {
                panic(err)
        }

        return string(template)
}

func (fr *Frontend) css() string {
        css, err := Asset("data/style.css")

        if err != nil {
                panic(err)
        }

        return string(css)
}

func (fr *Frontend) IndexPage() []byte {
        style := fr.css()
        js := fr.javascript()
        html := fr.html()

        type Variables struct {
                Css string
                Javascript string
        }

        variables := Variables{
                Css: style,
                Javascript: js,
        }

        buff := &bytes.Buffer{}

        tmpl, err := template.New("index.html").Parse(html)

        if err != nil {
                panic(err)
        }

        tmpl.Execute(buff, variables)

        return buff.Bytes()
}

func (fr *Frontend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        page := fr.IndexPage()

        w.Write(page)
}
