package main

import (
        "testing"
)

func Test_activeports(t *testing.T) {
        tests := mktests()

        expect := []int{80, 90}

        actual := tests.activeports()

        for i, acp := range actual {
                exp := expect[i]

                if acp != exp {
                        t.Error("Expected %v but received %v", exp, acp)
                }
        }

        // Repeat to test the cache.
        for i, acp := range actual {
                exp := expect[i]

                if acp != exp {
                        t.Error("Expected %v (in cache) but received %v", exp, acp)
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
                        t.Error("Expected %v but received %v", exp, acp)
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
                        t.Error("Expected %v but received %v", exp, acp)
                }
        }
}

func mkresults() *resultset {
        rset := &resultset{}
        rset.tests = mktests()

        return rset
}

func mktests() *testset {
        tests := &testset{}

        case80 := &testcase{}
        case80.port = 80

        case90 := &testcase{}
        case90.port = 90

        tests.cases = []*testcase{
                case80,
                case90,
        }

        return tests
}
