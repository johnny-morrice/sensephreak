package scanner

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "sync"
        "net/http"
        "github.com/pkg/errors"
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
        Id int
        host string
        apiport int
        sem chan struct{}
}

func Launch(hostname string, apiport, conns int) (*Scan, error) {
        newid := make(chan int)
        errch := make(chan error)

        scan := &Scan{}
        scan.host = hostname
        scan.apiport = apiport
        scan.sem = make(chan struct{}, conns)

        go func() {
                url := scan.Apipath("/test", apiport)

                resp, err := http.Post(url, jsontype, nilreader())

                if err != nil {
                        errch<- errors.Wrap(err, "Failed to launch test")

                        return
                }

                defer resp.Body.Close()

                var id *int
                dec := json.NewDecoder(resp.Body)
                err = dec.Decode(id)

                if err != nil {
                        errch<- errors.Wrap(err, "Failed to read id")

                        return
                }

                newid<- *id
        }();

        select {
        case id := <-newid:

                scan.Id = id
                return scan, nil
        case err := <-errch:
                return nil, err
        }
}

func (scan *Scan) Ping(port int) error {
        errch := make(chan error)

        go func() {
                scan.withlimit(func() {
                        url := scan.Apipath(fmt.Sprintf("/test/%v", scan.Id), port)

                        resp, err := http.Post(url, jsontype, nilreader())

                        if err != nil {
                                errch<- err
                        }

                        defer resp.Body.Close()
                })

                close(errch)
        }()

        err, ok := <-errch

        if !ok {
                // Was successful.
                return nil
        }

        return err
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
        scan.sem<- struct{}{}

        f()

        <-scan.sem
}

func (scan *Scan) Apipath(part string, port int) string {
        return fmt.Sprintf("http://%v:%v/api/%v", scan.host, port, part)
}

func nilreader() io.Reader {
        return &bytes.Buffer{}
}

const jsontype = ""
