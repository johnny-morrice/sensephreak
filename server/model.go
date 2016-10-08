package server

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

// resultset represents a running or completed test.
type resultset struct {
	tests   *testset
	startport int
	endport int
	passing []int
}

// success means the port has passed the test.
func (rset *resultset) success(port int) {
	rset.passing = append(rset.passing, port)
}

// failports returns the ports that fail the test.
func (rset *resultset) failports() []int {
	good := map[int]struct{}{}

	for _, port := range rset.passing {
		good[port] = struct{}{}
	}

	bad := []int{}

	for port := rset.startport; port <= rset.endport; port++ {
		if _, exempt := good[port]; exempt {
			continue
		}

		if _, present := rset.tests.portcache[port]; !present {
			continue
		}

		bad = append(bad, port)
	}

	return bad
}
