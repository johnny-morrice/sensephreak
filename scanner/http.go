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

func get(url string) (io.ReadCloser, error) {
	return nil, nil
}

func post(url, ctype string, content io.Reader) (io.ReadCloser, error) {
	return nil, nil
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
