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

type Scan struct {
	Id      int
	Host    string
	Apiport int
	Ports []int
	Conns int
	sem     chan struct{}
}

func (scan *Scan) Launch() error {
	scan.sem = make(chan struct{}, scan.Conns)

	url := scan.Apipath("/test", scan.Apiport)

	resp, err := http.Post(url, plaintype, nilreader())

	if err != nil {
		return errors.Wrap(err, "Failed to launch test")
	}

	defer resp.Body.Close()

	var id int
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&id)

	if err != nil {
		return errors.Wrap(err, "Failed to read id")
	}

        scan.Id = id

        return nil
}

func (scan *Scan) Scanall() ([]int, error) {
	wg := sync.WaitGroup{}

	wg.Add(len(scan.Ports))

	for _, p := range scan.Ports {
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
	url := scan.Apipath(fmt.Sprintf("/test/%v", scan.Id), scan.Apiport)

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
	return fmt.Sprintf("http://%v:%v/api%v", scan.Host, port, part)
}

func nilreader() io.Reader {
	return &bytes.Buffer{}
}

const plaintype = "text/plain"
