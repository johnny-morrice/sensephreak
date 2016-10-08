package server

import (
	"fmt"
	"log"
        "net"
	"net/http"
	"sync"
        "os"
        "github.com/gorilla/mux"
)

func Serve(bind net.IP, hostname string, ports []int) {
	ph := mkphreak(bind.String(), hostname)

        tests := make([]*testcase, len(ports))

        for i, p := range ports {
                tests[i] = ph.addtestcase(p)
        }

	wg := sync.WaitGroup{}

	wg.Add(len(tests))
	for _, tc := range tests {
		tcase := tc
		go func() {
			ph.servetest(tcase)

			wg.Done()
		}()
	}

	wg.Add(1)
	go func() {
		ph.serveweb()

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		ph.mainloop()

		wg.Done()
	}()

	wg.Wait()
}

func mkphreak(bind, hostname string) *phreak {
	ph := &phreak{}
	ph.commands = make(chan command)
	ph.tests = &testset{}
	ph.webport = Webport
	ph.bind = bind
        ph.hostname = hostname

	return ph
}

// phreak checks if your firewall is blocking you from seeing some ports.
type phreak struct {
	tests    *testset
	rsets    []*resultset
	commands chan command
	webport  int
	bind     string
        hostname string
}

// serveweb runs a webserver for the main API and web interface.
func (ph *phreak) serveweb() {
	srv := &http.Server{}
	srv.Addr = fmt.Sprintf("%v:%v", ph.bind, ph.webport)

	api := &phapi{}
	api.commands = ph.commands

	front := &frontend{}
	front.ports = ph.tests.activeports()
	front.host = ph.hostname
	front.apiport = Webport

        webtest := ph.addtestcase(Webport)
	r := mux.NewRouter()

	r.HandleFunc("/", front.index).Methods("GET")
        r.HandleFunc("/script.js", front.javascript).Methods("GET")
	r.HandleFunc("/jquery-3.1.1.min.js", front.jquery).Methods("GET")
        r.HandleFunc("/script.js.map", front.sourcemap).Methods("GET")
	r.HandleFunc("/api/test", api.newtest).Methods("POST")
	r.HandleFunc("/api/test/{resultset}", api.getresults).Methods("GET")
        r.Handle("/api/test/{resultset}/ping", webtest)

	srv.Handler = r

        if trace {
                loglisten(srv)
        }

	err := srv.ListenAndServe()

        fmt.Fprintf(os.Stderr, "%v\n", err)
}

// servetest runs a webserver on the given port for the /ping API call.
func (ph *phreak) servetest(tcase *testcase) {
	srv := &http.Server{}
	srv.Addr = fmt.Sprintf("%v:%v", ph.bind, tcase.port)

        if trace {
                loglisten(srv)
        }

	srv.Handler = tcase.handler()

	err := srv.ListenAndServe()

        fmt.Fprintf(os.Stderr, "%v\n", err)
}

// addtestcase adds a new test case to the set for the given port, and returns
// it.
func (ph *phreak) addtestcase(port int) *testcase {
	tcase := &testcase{}
	tcase.port = port
	tcase.set = ph.tests
	tcase.commands = ph.commands
        tcase.hostname = ph.hostname

	ph.tests.cases = append(ph.tests.cases, tcase)

	return tcase
}

// mailoop executes the main application logic loop.  Controller actions
// communicate with the loop over the `commands` channel.
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

// launch a new test.
func (ph *phreak) launch(r registration) {
	rset := &resultset{}
	rset.tests = ph.tests

	id := len(ph.rsets)
	ph.rsets = append(ph.rsets, rset)

	r.newid <- id
}

// ping the service to show you can access a port.
func (ph *phreak) ping(r result) error {
	if !ph.okresultid(r.set) {
		return fmt.Errorf("Bad result id: %v", r.set)
	}

	rset := ph.rsets[r.set]

	rset.success(r.port)

	r.done <- struct{}{}

	return nil
}

// badports responds to a query for the failing ports.
func (ph *phreak) badports(q query) error {
	if !ph.okresultid(q.rset) {
		close(q.failports)
		return fmt.Errorf("Bad result id: %v", q.rset)
	}

	rset := ph.rsets[q.rset]

	badports := rset.failports()

	q.failports <- badports

	return nil
}

func (ph *phreak) okresultid(resultset uint64) bool {
	return int(resultset) < len(ph.rsets)
}

type comtype uint8

const (
	_NEWTEST = iota
	_GETRESULT
	_PING
)

type command struct {
	ctype comtype
	reg   registration
	query query
	ping  result
}

type query struct {
	rset      uint64
	failports chan []int
}

type registration struct {
	newid chan int
}

type result struct {
	port int
	set  uint64
	done chan struct{}
}

func loglisten(srv *http.Server) {
        fmt.Fprintf(os.Stderr, "Serving on: %v\n", srv.Addr)
}

const Webport = 80
const debug = true
const trace = false
