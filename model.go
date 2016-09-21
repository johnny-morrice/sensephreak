package main

// testset is a specification of the ports that will be tested.
type testset struct {
	cases     []*testcase
	portcache []int
}

// activeports returns the ports that will be tested.
func (tset *testset) activeports() []int {
	if tset.portcache != nil {
		return tset.portcache
	}

	ports := []int{}

	for _, tc := range tset.cases {
		ports = append(ports, tc.port)
	}

	tset.portcache = ports

	return ports
}

// resultset represents a running or completed test.
type resultset struct {
	tests   *testset
	passing []int
}

// success means the port has passed the test.
func (rset *resultset) success(port int) {
	rset.passing = append(rset.passing, port)
}

// failports returns the ports that fail the test.
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
