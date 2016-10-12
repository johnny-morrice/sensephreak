package util

import (
        "testing"
)

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
        for p := Portmin; p <= Portmax; p++ {
                if _, bad := skip[p]; !bad {
                        ports = append(ports, p)
                }
        }

        return ports
}
