package server

import (
	"testing"
)

func Test_newuser(t *testing.T) {
	acc := &accounts{}

	users := []*user{
		acc.newuser(),
		acc.newuser(),
	}

	for i, expectid := range []int {0, 1} {
		actualid := users[i].id
		if actualid != expectid {
			t.Error("Expected", expectid, "but received", actualid)
		}
	}

	if len(users) != 2 {
		t.Error("Expected length 2 but received", len(users))
	}
}

func Test_getuser(t *testing.T) {
	acc := &accounts{}
	var user *user
	var err error

	for _, badid := range []int {-1, 0} {
		user, err = acc.getuser(badid)

		if user != nil && err != nil {
			t.Error("Expected nil user and non-nil error for id", badid)
		}
	}

	acc.newuser()

	user, err = acc.getuser(0)

	if user == nil || err != nil {
		t.Error("Expected non-nil user and nil error but got ", err)
	}
}

func Test_activeports(t *testing.T) {
        tests := mktests()

        for i, ports := range []map[int]struct{} {
		tests.activeports(),
		tests.activeports(),
	} {
		if len(ports) != 3 {
			t.Error("(case", i, ") Expected length 3 but was ", ports)

			return
		}

		_, ok80 := ports[80]
		_, ok90 := ports[90]
		_, ok91 := ports[91]

		if !(ok80 && ok90 && ok91) {
			t.Error("(case", i, ") Expected ports but received", ports)
		}
	}


}

func Test_success(t *testing.T) {
	rset := mkresults()

	rset.success(90)

	expect := []int{80}

	actual := rset.failports()

	for i, acp := range actual {
		exp := expect[i]

		if acp != exp {
			t.Error("Expected", exp, "but received", acp)
		}
	}
}

func Test_failports(t *testing.T) {
	rset := mkresults()

	expect := []int{80, 90}

	actual := rset.failports()

	for i, acp := range actual {
		exp := expect[i]

		if acp != exp {
			t.Error("Expected", exp, "but received", acp)
		}
	}
}

func Test_Ports(t *testing.T) {
	formats := []string {
		"-10",
		"-10:20",
		"-10:20,+15",
		"+10",
		"+10:20",
		"+10:20,-15",
	}

	expected := [][]int {
		allbut([]int{10}),
		allbut([]int{10, 11, 12, 13, 14, 15, 16, 17,  18, 19, 20}),
		allbut([]int{10, 11, 12, 13, 14,     16, 17,  18, 19, 20}),
		[]int{10},
		[]int{10, 11, 12, 13, 14, 15, 16, 17,     19, 20},
		[]int{10, 11, 12, 13, 14,     16, 17,     19, 20},
	}

	for i, form := range formats {
		exp := expected[i]

		ports, err := Ports(form, 18)

                if err != nil {
                        t.Error(err)
                } else {
                        if len(exp) != len(ports) {
                                unexpectedports(t, i, exp, ports)
                                continue
                        }

                        for j, actport := range ports {
                                export := exp[j]
                                if actport != export {
                                        unexpectedports(t, i, exp, ports)
                                        break
                                }
                        }
                }
	}
}

func unexpectedports(t *testing.T, i int, exp, ports []int) {
        const limit = 10
        errcnt := 0
        for j, ep := range exp {
                ap := ports[j]

                if errcnt < limit {
                        if ep != ap {
                                t.Error("At test", i, "index", j, "expected", ep, "received", ap)
                        }

                        errcnt++
                } else {
                        break
                }
        }
}

func allbut(skiplist []int) []int {
        skip := map[int]struct{}{}

        for _, p := range skiplist {
                skip[p] = struct{}{}
        }

        var ports []int
        for p := portmin; p <= portmax; p++ {
                if _, bad := skip[p]; !bad {
                        ports = append(ports, p)
                }
        }

        return ports
}

func mkresults() *resultset {
	rset := &resultset{}
	rset.tests = mktests()
	rset.startport = 80
	rset.endport = 90

	return rset
}

func mktests() *testset {
	tests := &testset{}

	case80 := &testcase{}
	case80.port = 80

	case90 := &testcase{}
	case90.port = 90

	case91 := &testcase{}
	case91.port = 91

	tests.cases = []*testcase{
		case80,
		case90,
		case91,
	}

	return tests
}
