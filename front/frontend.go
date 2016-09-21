//go:generate go-bindata -pkg front data/
package front

import (
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

func (fr *Frontend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        page := fr.html()

        w.Write([]byte(page))
}
