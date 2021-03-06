package server

import (
        "fmt"

        "github.com/johnny-morrice/sensephreak/util"
)

// testset is a specification of the ports that will be tested.
type testset struct {
        cases     []*testcase
        portcache map[int]struct{}
}

func (tset *testset) activeports() map[int]struct{} {
        if tset.portcache == nil {
                tset.portcache = make(map[int]struct{})
                for _, tc := range tset.cases {
                        tset.portcache[tc.port] = struct{}{}
                }
        }

        return tset.portcache
}

type accounts struct {
        users []*user
}

func (a *accounts) getuser(id int) (*user, error) {
        if id < 0 || id >= len(a.users) {
                return nil, fmt.Errorf("Invalid user id: %v", id)
        }

        user := a.users[id]

        if user.id != id {
                panic("Corrupt accounts db")
        }

        return user, nil
}

func (a *accounts) newuser() *user {
        id := len(a.users)
        user := &user{id: id}
        a.users = append(a.users, user)

        return user
}

type user struct {
        id int
}

// resultset represents a running or completed test.
type resultset struct {
	tests   *testset
	startport int
	endport int
	passing []int

        user *user
}

// success means the port has passed the test.
func (rset *resultset) success(port int) {
	rset.passing = append(rset.passing, port)
}

// failports returns the ports that fail the test.
func (rset *resultset) failports() []util.PortStatus {
	good := map[int]struct{}{}

	for _, port := range rset.passing {
		good[port] = struct{}{}
	}

	var out []util.PortStatus
	active := rset.tests.activeports()

	for port := rset.startport; port <= rset.endport; port++ {
                status := util.PortStatus{}
                status.Port = port

		if _, exempt := good[port]; exempt {
			status.State = util.PortOk
		} else if _, present := active[port]; !present {
			status.State = util.PortOmitted
		} else {
                        status.State = util.PortBlocked
                }

		out = append(out, status)
	}

	return out
}


const nouser = -1
