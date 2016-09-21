package main

import (
        "fmt"
        "log"
        "net/http"
        "sync"
        "github.com/gorilla/mux"
)

func main() {
        ports := makeportlist()

        ph := &phreak{}
        ph.commands = make(chan command)
        ph.tests = &testset{}
        ph.webport = webport
        ph.bind = bindinter

        wg := sync.WaitGroup{}

        wg.Add(1)
        go func() {
                ph.serveweb()

                wg.Done()
        }()

        wg.Add(len(ports))
        for _, p := range ports {
                go func() {
                        ph.servetest(p)

                        wg.Done()
                }()
        }

        wg.Add(1)
        go func() {
                ph.mainloop()

                wg.Done()
        }()

        wg.Wait()
}

func basicskip() map[int]struct{} {
        skiplist := map[int]struct{}{}

        for i := 1; i < sysportmax; i++ {
                skiplist[i] = struct{}{}
        }

        return skiplist
}

func makeportlist() []int {
        skiplist := basicskip()

        ports := []int{}

        for i := 1; i < portmax; i++ {
                if _, skip := skiplist[i]; skip {
                        continue
                }

                ports = append(ports, i)
        }

        return ports
}

type phreak struct {
        tests *testset
        rsets []*resultset
        commands chan command
        webport int
        bind string
}

func (ph *phreak) serveweb() {
        srv := &http.Server{}
        srv.Addr = fmt.Sprintf("%v:%v", ph.bind, ph.webport)

        api := &phapi{}
        api.commands = ph.commands

        r := mux.NewRouter()
        r.HandleFunc("/api/test", api.newtest).Methods("POST")
        r.HandleFunc("/api/test/{resultset}", api.getresults).Methods("GET")

        srv.Handler = r

        srv.ListenAndServe()
}

func (ph *phreak) servetest(port int) {
        srv := &http.Server{}
        srv.Addr = fmt.Sprintf("%v:%v", ph.bind, port)

        tcase := ph.addtestcase(port)

        r := mux.NewRouter()
        r.Handle("/api/test/{resultset}/ping", tcase).Methods("POST")

        srv.Handler = r

        srv.ListenAndServe()
}

func (ph *phreak) addtestcase(port int) *testcase {
        tcase := &testcase{}
        tcase.port = port
        tcase.set = ph.tests
        tcase.commands = ph.commands

        ph.tests.cases = append(ph.tests.cases, tcase)

        return tcase
}

func (ph *phreak) mainloop() {
        for cmd := range ph.commands {
                var err error

                switch cmd.ctype {
                case _NEWTEST:
                        ph.launch(cmd.reg)
                case _PING:
                        err = ph.ping(cmd.ping)
                case _GETRESULT:
                        err = ph.badports(cmd.query)
                }

                if err != nil {
                        log.Printf("Error in mainloop: %v", err)
                        err = nil
                }
        }
}

func (ph *phreak) ping(r *result) error {
        if !ph.okresultid(r.resultset) {
                return fmt.Errorf("Bad result id: %v", r.resultset)
        }

        rset := ph.rsets[r.resultset]

        rset.pass(r.port)

        return nil
}

func (ph *phreak) okresultid(resultset uint64) bool {
        return int(resultset) < len(ph.rsets)
}

func (ph *phreak) launch(r *registration) {
        rset := &resultset{}
        rset.tests = ph.tests

        id := len(ph.rsets)
        ph.rsets = append(ph.rsets, rset)

        r.newid<- id
}

func (ph *phreak) badports(q *query) error {
        if !ph.okresultid(q.rset) {
                close(q.failports)
                return fmt.Errorf("Bad result id: %v", q.rset)
        }

        rset := ph.rsets[q.rset]

        badports := rset.failports()

        q.failports<- badports

        return nil
}

type comtype uint8
const (
        _NEWTEST = iota
        _GETRESULT
        _PING
)

type command struct {
        ctype comtype
        reg *registration
        query *query
        ping *result
}

const webport = 80
const bindinter = "0.0.0.0"
const sysportmax = 1000
const portmax = 65536
