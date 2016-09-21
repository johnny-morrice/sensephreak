package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"sync"
)

func Scanall(hostname string, apiport int, ports []int) ([]int, error) {
	scan, err := Launch(hostname, apiport, 50)

	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}

	wg.Add(len(ports))

	for _, p := range ports {
		port := p
		go func() {
			scan.Ping(port)
			wg.Done()
		}()
	}
	wg.Wait()

	failed, err := scan.BadPorts()

	if err != nil {
		return nil, err
	}

	return failed, nil
}

type Scan struct {
	Id      int
	host    string
	apiport int
	sem     chan struct{}
}

func Launch(hostname string, apiport, conns int) (*Scan, error) {
	scan := &Scan{}
	scan.host = hostname
	scan.apiport = apiport
	scan.sem = make(chan struct{}, conns)

	url := scan.Apipath("/test", apiport)

	resp, err := http.Post(url, plaintype, nilreader())

	if err != nil {
		return nil, errors.Wrap(err, "Failed to launch test")
	}

	defer resp.Body.Close()

	var id int
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&id)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to read id")
	}

        scan.Id = id

        return scan, nil
}

func (scan *Scan) Ping(port int) error {
        var ret error

	scan.withlimit(func() {
		url := scan.Apipath(fmt.Sprintf("/test/%v/ping", scan.Id), port)

		resp, err := http.Post(url, plaintype, nilreader())

		if err != nil {
			ret = err

			return
		}

		defer resp.Body.Close()
	})

	return ret
}

func (scan *Scan) BadPorts() ([]int, error) {
	url := scan.Apipath(fmt.Sprintf("/test/%v", scan.Id), scan.apiport)

	resp, err := http.Get(url)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to GET BadPorts")
	}

	defer resp.Body.Close()

	ports := []int{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&ports)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode JSON in BadPorts")
	}

	return ports, nil
}

func (scan *Scan) withlimit(f func()) {
	scan.sem <- struct{}{}

	f()

	<-scan.sem
}

func (scan *Scan) Apipath(part string, port int) string {
	return fmt.Sprintf("http://%v:%v/api%v", scan.host, port, part)
}

func nilreader() io.Reader {
	return &bytes.Buffer{}
}

const plaintype = "text/plain"
