// +build js

package scanner

import (
	"io"
	"io/ioutil"
	"strings"

	"honnef.co/go/js/xhr"
)

func initcookies() {
	// Not needed with XHR.
}

func get(url string) (io.ReadCloser, int, error) {
	r := xhr.NewRequest("GET", url)
	r.WithCredentials = true

	err := r.Send(nilreader)

	return responsedata(r, err)
}

func post(url, ctype string, content io.Reader) (io.ReadCloser, int, error) {
	r := xhr.NewRequest("POST", url)
	r.WithCredentials = true
	r.SetRequestHeader("Content-Type", ctype)

	data, err := ioutil.ReadAll(content)

	if err != nil {
		return nil, 0, err
	}

	err = r.Send(data)

	return responsedata(r, err)
}

func responsedata(r *xhr.Request, err error) (io.ReadCloser, int, error) {
	if err == nil {
		reader := strings.NewReader(r.ResponseText)
		closer := ioutil.NopCloser(reader)
		return closer, r.Status, nil
	} else {
		return nil, 0, err
	}
}

func dumpCookies(when,  urltext string) {
	// Not supported with XHR.
}
