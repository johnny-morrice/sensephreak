// +build !js

package scanner

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/pkg/errors"
)

func initcookies() {
	jar, err := cookiejar.New(nil)

	if err != nil {
		panic(errors.Wrap(err, "Failed to create cookiejar"))
	}

	http.DefaultClient.Jar = jar
}

func get(url string) (io.ReadCloser, int, error) {
	resp, err := http.Get(url)
	return responsedata(resp, err)
}

func post(url, ctype string, content io.Reader) (io.ReadCloser, int, error) {
	resp, err := http.Post(url, ctype, content)
	return responsedata(resp, err)
}

func responsedata(resp *http.Response, err error) (io.ReadCloser, int, error) {
	if err == nil {
		return resp.Body, resp.StatusCode, nil
	} else {
		return nil, 0, err
	}
}

func dumpCookies(when,  urltext string) {
	u, err := url.Parse(urltext)

	if err != nil {
		panic(err)
	}

	cookies := http.DefaultClient.Jar.Cookies(u)

	for _, c := range cookies {
		fmt.Fprintf(os.Stderr, "Cookie at %v: %v\n", when, c)
	}

	if len(cookies) == 0 {
		fmt.Fprintf(os.Stderr, "No cookies at %v\n", when)
	}
}
