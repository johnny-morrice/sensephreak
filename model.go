package main


type testset struct {
        cases []*testcase
        portcache []int
}

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

type query struct {
        rset uint64
        failports chan []int
}

type registration struct {
        newid chan int
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
