package server

import (
	"fmt"
        "net"
	"net/http"
	"sync"
        "os"
        "github.com/gorilla/mux"
        "github.com/gorilla/sessions"

        "github.com/johnny-morrice/sensephreak/util"
)

type Server struct {
        Bind net.IP
        Hostname string
        Webport int
        Ports []int
        Secret string

        // For frontend
        Title string
        Heading string
}

func (s Server) Serve() {
        globalSessionStore = sessions.NewCookieStore([]byte(s.Secret))

	ph := mkphreak(s)

        tests := make([]*testcase, len(s.Ports))

        for i, p := range s.Ports {
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

func mkphreak(s Server) *phreak {
	ph := &phreak{}
	ph.commands = make(chan command)
	ph.tests = &testset{}
        ph.accounts = &accounts{}
	ph.webport = s.Webport
	ph.bind = s.Bind.String()
        ph.hostname = s.Hostname

        ph.front = &frontend{}
        ph.front.host = s.Hostname
        ph.front.title = s.Title
        ph.front.heading = s.Heading
        ph.front.apiport = s.Webport

	return ph
}

// phreak checks if your firewall is blocking you from seeing some ports.
type phreak struct {
        accounts *accounts
	tests    *testset
	rsets    []*resultset
	commands chan command
	webport  int
	bind     string
        hostname string
        front    *frontend
}

// serveweb runs a webserver for the main API and web interface.
func (ph *phreak) serveweb() {
	srv := &http.Server{}
	srv.Addr = fmt.Sprintf("%v:%v", ph.bind, ph.webport)

	api := &phapi{}
	api.commands = ph.commands

        webtest := ph.addtestcase(ph.webport)
	r := mux.NewRouter()

	r.HandleFunc("/", ph.front.index).Methods("GET")
        r.HandleFunc("/script.js", ph.front.javascript).Methods("GET")
	r.HandleFunc("/jquery-3.1.1.min.js", ph.front.jquery).Methods("GET")
        r.HandleFunc("/script.js.map", ph.front.sourcemap).Methods("GET")
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
		switch cmd.ctype {
		case _NEWTEST:
			ph.launch(cmd.reg)
		case _PING:
			ph.ping(cmd.ping)
		case _GETRESULT:
			ph.badports(cmd.query)
		}
	}
}

// launch a new test.
func (ph *phreak) launch(r registration) {
        var err error
        var user *user
        if r.userid == nouser {
                user = ph.accounts.newuser()
        } else {
                user, err = ph.accounts.getuser(r.userid)
        }

        reply := regisreply{}
        reply.err = err

        if err == nil {
                rset := &resultset{}
                rset.tests = ph.tests
                rset.startport = r.StartPort
                rset.endport = r.EndPort
                rset.user = user

                id := len(ph.rsets)
                ph.rsets = append(ph.rsets, rset)

                reply.scanid = id
                reply.userid = user.id
        }

	r.reply <- reply
}

// ping the service to show you can access a port.
func (ph *phreak) ping(r ping) error {
        reply := pingreply{}

	if !ph.okresultid(r.set) {
		reply.err = fmt.Errorf("Bad result id: %v", r.set)
	}

        var user *user
        user, reply.err = ph.accounts.getuser(r.userid)

        if reply.err == nil {
        	rset := ph.rsets[r.set]

                if rset.user == user {
                        rset.success(r.port)

                } else {
                        reply.err = forbidden(user)
                }
        }

	r.reply <- reply

	return reply.err
}

// badports responds to a query for the failing ports.
func (ph *phreak) badports(q query) error {
        reply := queryreply{}

	if !ph.okresultid(q.rset) {
		reply.err = fmt.Errorf("Bad result id: %v", q.rset)
	}

        var user *user
        user, reply.err = ph.accounts.getuser(q.userid)

        if reply.err == nil {
                rset := ph.rsets[q.rset]

                if rset.user == user {
                        badports := rset.failports()

                        reply.portinfo = badports
                } else {
                        reply.err = forbidden(user)
                }
        }

	q.reply <- reply

	return reply.err
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
	ping  ping
}

type userdata struct {
        userid int
}

type queryreply struct {
        portinfo []util.PortStatus
        err error
}

type query struct {
        userdata
	rset      uint64
	reply chan queryreply
}

type regisreply struct {
        userdata
        scanid int
        err error
}

type registration struct {
	util.LaunchData
        userdata
	reply chan regisreply
}

type pingreply struct {
        err error
}

type ping struct {
        userdata
	port int
	set  uint64
	reply chan pingreply
}

func loglisten(srv *http.Server) {
        fmt.Fprintf(os.Stderr, "Serving on: %v\n", srv.Addr)
}

func forbidden(user *user) error {
        return fmt.Errorf("Forbidden for user: %v")
}

const debug = false
const trace = false
