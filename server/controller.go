package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

        "github.com/johnny-morrice/ctrl"
        "github.com/gorilla/mux"
        "github.com/gorilla/sessions"

        "github.com/johnny-morrice/sensephreak/util"
)


// testcase is a controller that runs on the given port.
type testcase struct {
	port     int
	set      *testset
	commands chan<- command
	hostname string
}

func (tcase *testcase) handler() http.Handler {
	r := mux.NewRouter()
	r.Handle("/api/test/{resultset}/ping", tcase)

        // Use CORS to allow all origins.
        handler := newcorshandler(r, fmt.Sprintf("http://%v", tcase.hostname))

        return handler
}

func (tc *testcase) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Forward declaration for goto.
	ping := ping{}
	cmd := command{}
        reply := pingreply{}
        var session *sessions.Session
        var userid int

	nocache(w)

	c := ctrl.New(w, r)

	rset, err := resultsetparam(c)

	if err != nil {
		goto ERROR
	}

        session, err = getsession(r)

        if err != nil {
                log.Printf("Session error: %v", err)
        }

        userid = getsessionuser(session)

        // Send command to main loop.
	ping.port = tc.port
	ping.set = rset
        ping.reply = make(chan pingreply)
        ping.userid = userid

	cmd.ctype = _PING
	cmd.ping = ping

	tc.commands<- cmd

	// Receive command from main loop.
	reply = <-ping.reply
        err = reply.err

        if err != nil {
                goto ERROR
        }

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

// phapi is the API controller.
type phapi struct {
	commands chan<- command
}

func (api *phapi) getresults(w http.ResponseWriter, r *http.Request) {
	// Forward declare variables because we are using goto
	q := query{}
	cmd := command{}
        reply := queryreply{}
        var session *sessions.Session
        var userid int

	nocache(w)

	if debug {
		fmt.Fprintln(os.Stderr, "getresults")
	}

	c := ctrl.New(w, r)

	resultset, err := resultsetparam(c)

	if err != nil {
		goto ERROR
	}

        session, err = getsession(r)

        if err != nil {
		log.Printf("Session error: %v", err)
        }

        userid = getsessionuser(session)

        // Send command to main loop.
	q.rset = resultset
	q.reply = make(chan queryreply)
        q.userid = userid

	cmd.ctype = _GETRESULT
	cmd.query = q

	api.commands<- cmd

        // Receive command from main loop.
	reply = <-q.reply
        err = reply.err

        if err != nil {
                goto ERROR
        }

	err = c.ServeJson(reply.badports)

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

        var userid int
        var session *sessions.Session
        reg := registration{}
        cmd := command{}
        reply := regisreply{}

	nocache(w)

        packet := &util.LaunchData{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(packet)


	if err != nil {
		goto ERROR
	}

        session, err = getsession(r)

        if err != nil {
		log.Printf("Session error: %v", err)
        }

        userid = getsessionuser(session)

        // Send command to main loop.
	reg.reply = make(chan regisreply)
	reg.LaunchData = *packet
        reg.userid = userid

	cmd.ctype = _NEWTEST
	cmd.reg = reg

	api.commands <- cmd

        // Receive reply from main loop.
	reply = <-reg.reply
        err = reply.err

        if err != nil {
                goto ERROR
        }

        if userid != reply.userid {
                setsessionuser(session, reply.userid)

                err = session.Save(r, w)

                if err != nil {
                        goto ERROR
                }
        }

	err = c.ServeJson(reply.scanid)

ERROR:
	if err != nil {
		log.Printf("Error serving newtest: %v", err)

		c.InternalError()
	}
}

type corshandler struct {
	handler http.Handler
	allowed map[string]struct{}
	headers []string
}

func newcorshandler(h http.Handler, origins... string) http.Handler {
	ch := corshandler{}
	ch.handler = h
	ch.headers = []string{"Content-Type"}

	ch.allowed = map[string]struct{}{}
	for _, o := range origins {
		ch.allowed[o] = struct{}{}
	}

	return ch
}

func (ch corshandler) cors(w http.ResponseWriter, req *http.Request) {
	originheads := req.Header["Origin"]

	var origin string
	if len(originheads) == 1 {
		origin = originheads[0]
	} else {
		if trace {
                	log.Printf("Bad Origin header: %v", originheads)
		}

		return
        }

        _, any := ch.allowed["*"];
        _, ok := ch.allowed[origin];

        if any || ok {
		w.Header()["Access-Control-Allow-Credentials"] = []string{"true"}
		w.Header()["Access-Control-Allow-Origin"] = []string{origin}
		w.Header()["Access-Control-Allow-Headers"] = ch.headers
	}
}

func (ch corshandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ch.cors(w, req)

	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
	} else {
		ch.handler.ServeHTTP(w, req)
        }
}

func resultsetparam(c ctrl.C) (uint64, error) {
	resultset, err := c.GetMuxVar("resultset")

	if err != nil {
		return 0, err
	}

	setid, err := strconv.ParseUint(resultset, 10, 64)

	if err != nil {
		return 0, err
	}

	return setid, nil
}

var globalSessionStore *sessions.CookieStore

func getsession(r *http.Request) (*sessions.Session, error) {
        return globalSessionStore.Get(r, "session-name")
}

func getsessionuser(session *sessions.Session) int {
        maybeid, ok := session.Values["user"]

	if debug {
		log.Printf("session.Values: %v", session.Values)
		log.Printf("user id is: %v", maybeid)
	}

        if ok {
                var id int
                id, ok = maybeid.(int)

                if ok {
                    return id
                }
        }

        return nouser
}

func setsessionuser(session *sessions.Session, userid int) {
        session.Values["user"] = userid
}

func nocache(w http.ResponseWriter) {
	headnames := []string {
		"Cache-Control",
		"Pragma",
		"Expires",
	}

	headerval := [][]string {
		[]string{"no-cache", "no-store", "must-revalidate"},
		[]string{"no-cache"},
		[]string{"0"},
	}

	for i, title := range headnames {
		val := headerval[i]

		w.Header()[title] = val
	}
}
