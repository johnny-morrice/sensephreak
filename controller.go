package main

import (
	"fmt"
	"github.com/johnny-morrice/ctrl"
	"log"
	"net/http"
	"strconv"
)

// testcase is a controller that runs on the given port.
type testcase struct {
	port     int
	set      *testset
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
	// Forward declaration for goto.
	res := &result{}
	cmd := command{}

	c := ctrl.New(w, r)

	rset, err := resultsetparam(c)

	if err != nil {
		goto ERROR

	}

	res.port = tc.port
	res.set = rset

	cmd.ctype = _PING
	cmd.ping = res

	tc.commands <- cmd

	// Wait to be handled
	<-res.done

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

	api.commands <- cmd

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

	api.commands <- cmd

	id := <-reg.newid

	err := c.ServeJson(id)

	if err != nil {
		log.Printf("Error in newtest: %v", err)

		c.InternalError()
	}
}
