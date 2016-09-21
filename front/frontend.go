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

func (fr *Frontend) getasset(name string) string {
        template, err := Asset(name)

        if err != nil {
                panic(err)
        }

        return string(template)
}

func (fr *Frontend) javascript() string {
        return fr.getasset("data/script.js")
}

func (fr *Frontend) html() string {
        return fr.getasset("data/index.html")
}

func (fr *Frontend) css() string {
        return fr.getasset("data/style.css")
}

func (fr *Frontend) IndexPage() []byte {
        style := fr.css()
        js := fr.javascript()
        html := fr.html()

        type Variables struct {
                Css string
                Javascript string
                Ports []int
        }

        variables := Variables{
                Css: style,
                Javascript: js,
                Ports: fr.Ports,
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
