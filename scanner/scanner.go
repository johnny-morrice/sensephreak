package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/johnny-morrice/sensephreak/server"
)

type Scan struct {
	Id      int
	Host    string
	Apiport int
	StartPort int
	EndPort int
	Conns int
	Verbose bool

	sem     chan struct{}
}

func (scan *Scan) GoodPorts(badports []int) []int {
	badmap := make(map[int]struct{})
	var goodports []int

	for _, bad := range badports {
		badmap[bad] = struct{}{}
	}

	for p := scan.StartPort; p <= scan.EndPort; p++ {
		if _, bad := badmap[p]; !bad {
			goodports = append(goodports, p)
		}
	}

	return goodports
}

func (scan *Scan) Launch() error {
        if scan.Conns == 0 {
                scan.Conns = DefaultConns
        }

	initcookies()

	scan.sem = make(chan struct{}, scan.Conns)

	url := scan.Apipath("/test", scan.Apiport)

	packet := server.LaunchData{}
	packet.StartPort = scan.StartPort
	packet.EndPort = scan.EndPort

	buff := &bytes.Buffer{}
	enc := json.NewEncoder(buff)

	err := enc.Encode(packet)

	if err != nil {
		return errors.Wrap(err, "Failed to encode LaunchData")
	}

	if trace {
		dumpCookies("pre-launch", url)
	}

	var resp *http.Response
	resp, err = http.Post(url, plaintype, buff)

	if err != nil {
		return errors.Wrap(err, "Failed to launch test")
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Failed to launch test with bad response status: %v", resp.StatusCode)

		return err
	}

	if trace {
		dumpCookies("post-launch", url)
	}

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

	for p := scan.StartPort; p <= scan.EndPort; p++ {
		wg.Add(1)
		go func(port int) {
			err := scan.Ping(port)

			if scan.Verbose && err != nil {
				fmt.Fprintf(os.Stderr, "error pinging port %v: %v\n", port, err)
			}

			wg.Done()
		}(p)
	}
	wg.Wait()

	failed, err := scan.BadPorts()

	if err != nil {
		return nil, err
	}

	return failed, nil
}

func (scan *Scan) Ping(port int) error {
        var err error

	scan.withlimit(func() {
		url := scan.Apipath(fmt.Sprintf("/test/%v/ping", scan.Id), port)

		if trace {
			dumpCookies("pre-ping", url)
		}

		var resp *http.Response
		resp, err = http.Post(url, plaintype, nilreader())

		if err != nil {
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			err = fmt.Errorf("Ping failed with bad response status: %v", resp.StatusCode)

			return
		}
	})

	return err
}

func (scan *Scan) BadPorts() ([]int, error) {
	url := scan.Apipath(fmt.Sprintf("/test/%v", scan.Id), scan.Apiport)

	resp, err := http.Get(url)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to GET BadPorts")
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("BadPorts failed with response status: %v", resp.StatusCode)

		return nil, err
	}

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
const DefaultConns = 50
const trace = false
