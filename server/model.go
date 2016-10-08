package server

// testset is a specification of the ports that will be tested.
type testset struct {
	cases     []*testcase
	portcache []int
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
		if _, ok := good[port]; ok {
			continue
		}

		bad = append(bad, port)
	}

	return bad
}
