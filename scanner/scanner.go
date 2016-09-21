package main

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "sync"
        "net/http"
        "github.com/pkg/errors"
        "github.com/gopherjs/gopherjs/js"
)

func main() {
	js.Global.Set("Scanall", Scanall)
}

func Scanall(hostname string, apiport int, ports []int) error {
        scan, err := Launch(hostname, apiport, 50)

        if err != nil {
                return err
        }

        wg := sync.WaitGroup{}

        errch := make(chan error)
        wg.Add(len(ports))
        for _, p := range ports {
                port := p
                go func() {
                        err := scan.Ping(port)

                        errch<- err
                        wg.Done()
                }()
        }

        buff := &bytes.Buffer{}
        for err := range errch {
                buff.WriteString(err.Error())
                buff.WriteString("\n")
        }

        msg := buff.String()

        wg.Wait()

        if msg != "" {
                return errors.New(msg)
        }

        return nil
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
