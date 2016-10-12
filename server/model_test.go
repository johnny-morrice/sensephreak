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
