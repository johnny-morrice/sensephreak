package server

import (
	"testing"
)

func Test_activeports(t *testing.T) {
        tests := mktests()

        expect := []int{80, 90, 91}

        actualMap := tests.activeports()
	var actual []int

	for i, _ := range actualMap {
		actual = append(actual, i)
	}

        for i, acp := range actual {
                exp := expect[i]

                if acp != exp {
                        t.Error("Expected", exp, "but received", acp)
                }
        }

	actualMap = tests.activeports()
	actual = nil

	for i, _ := range actualMap {
		actual = append(actual, i)
	}

        // Repeat to test the cache.
        for i, acp := range actual {
                exp := expect[i]

                if acp != exp {
                        t.Error("(cached) Expected", exp, " but received", acp)
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
