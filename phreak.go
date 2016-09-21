package main

import (
        "fmt"
        "log"
        "net/http"
        "strconv"
        "sync"
        "github.com/gorilla/mux"
        "github.com/johnny-morrice/ctrl"
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

type phapi struct {
        commands chan<- command
}

func (api *phapi) getresults(w http.ResponseWriter, r *http.Request) {
        // Forward declare variables because we are using goto
        q := &query{}
        cmd := command{}
        var badports []int
        var ok bool

        c := ctrl.New(w, r)

        resultset, err := resultsetparam(c)

        if err != nil {
                goto ERROR
        }

        q.rset = resultset
        q.failports = make(chan []int)

        cmd.ctype = _GETRESULT
        cmd.query = q

        api.commands<- cmd

        badports, ok = <-q.failports

        if !ok {
                err = fmt.Errorf("Return channel closed")

                goto ERROR
        }

        err = c.ServeJson(badports)

        if err == nil {
                return
        }

ERROR:
        if err != nil {
                log.Printf("Error in getresults: %v", err)

                c.InternalError()
        }
}

func (api *phapi) newtest(w http.ResponseWriter, r *http.Request) {
        c := ctrl.New(w, r)

        reg := &registration{}
        reg.newid = make(chan int)

        cmd := command{}
        cmd.ctype = _NEWTEST
        cmd.reg = reg

        api.commands<- cmd

        id := <-reg.newid

        err := c.ServeJson(id)

        if err != nil {
                log.Printf("Error in newtest: %v", err)

                c.InternalError()
        }
}

type query struct {
        rset uint64
        failports chan []int
}

type registration struct {
        newid chan int
}

type testcase struct {
        port int
        set *testset
        commands chan<- command
}

func resultsetparam(c ctrl.C) (uint64, error) {
        resultset, err := c.Var("resultset")

        if err != nil {
                return 0, err
        }

        setid, err := strconv.ParseUint(resultset, 10, 64)

        if err != nil {
                return 0, err
        }

        return setid, nil
}

func (tc *testcase) ServeHTTP(w http.ResponseWriter, r *http.Request) {
        c := ctrl.New(w, r)

        rset, err := resultsetparam(c)

        if err != nil {
                goto ERROR

        }

        go func() {
                r := &result{}
                r.port = tc.port
                r.resultset = rset

                cmd := command{}
                cmd.ctype = _PING
                cmd.ping = r

                tc.commands<- cmd
        }()

        err = c.ServeJson(true)

        if err == nil {
                return
        }

ERROR:
        if err != nil {
                log.Printf("Error in testcase handler: %v", err)

                c.InternalError()
        }
}

type testset struct {
        cases []*testcase
}

// TODO could be cached.
func (tset *testset) activeports() []int {
        ports := []int{}

        for _, tc := range tset.cases {
                ports = append(ports, tc.port)
        }

        return ports
}

type resultset struct {
        tests *testset
        passing []int
}

func (rset *resultset) pass(port int) {
        rset.passing = append(rset.passing, port)
}

func (rset *resultset) failports() []int {
        active := rset.tests.activeports()

        good := map[int]struct{}{}

        for _, port := range rset.passing {
                good[port] = struct{}{}
        }

        bad := []int{}

        for _, port := range active {
                if _, ok := good[port]; ok {
                        continue
                }

                bad = append(bad, port)
        }

        return bad
}

type result struct {
        port int
        resultset uint64
}

const webport = 80
const bindinter = "0.0.0.0"
const sysportmax = 1000
const portmax = 65536
