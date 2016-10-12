package util

import (
        "fmt"
        "sort"
        "strings"
        "strconv"

        "github.com/pkg/errors"
)

type portcommand uint8

const (
        addport = portcommand(iota)
        removeport
)

type portformat struct {
        cmd portcommand
        startport int
        endport int
}

func (pf portformat) transform(ports []int) []int {
        if pf.cmd == addport {
                for p := pf.startport; p <= pf.endport; p++ {
                        ports = append(ports, p)
                }

                return ports
        } else {
                filter := map[int]struct{}{}

                for p := pf.startport; p <= pf.endport; p++ {
                        filter[p] = struct{}{}
                }

                newports := []int{}

                for _, p := range ports {
                        if _, bad := filter[p]; !bad {
                                newports = append(newports, p)
                        }
                }

                return newports
        }
}

func parsesingleformat(format string) (portformat, error) {
        pf := portformat{}

        switch format[0:1] {
        case "+":
                pf.cmd = addport
        case "-":
                pf.cmd = removeport
        default:
                return pf, fmt.Errorf("Unexpected format command: %v", format[0:1])
        }

        format = format[1:]

        parts := strings.Split(format, ":")

        var err error
        pf.startport, err = strconv.Atoi(parts[0])

        if err != nil {
                return pf, errors.Wrap(err, "Could not parse startport")
        }

        if len(parts) == 2 {
                pf.endport, err = strconv.Atoi(parts[1])

                if err != nil {
                         return pf, errors.Wrap(err, "Could not parse endport")
                }
        } else if len(parts) == 1 {
                pf.endport = pf.startport
        } else {
                return pf, fmt.Errorf("Unexpected length: %v", len(parts))
        }

        return pf, nil
}

func parseportformat(format string) ([]portformat, error) {
        singles := strings.Split(format, ",")

        out := make([]portformat, len(singles))

        for i, form := range singles {
                p, err := parsesingleformat(form)

                if err != nil {
                        return nil, errors.Wrap(err, fmt.Sprintf("Error parsing format #%v", i))
                }

                out[i] = p
        }

        return out,nil
}

func Ports(format string, webport int) ([]int, error) {
        pfs, err := parseportformat(format)

        if err != nil {
                return nil, err
        }

        var ports []int

        if len(pfs) < 1 || pfs[0].cmd != addport {
                ports = initports(Portmin, Portmax, webport)
        } else {
                ports = initports(pfs[0].startport, pfs[0].endport, webport)
                pfs = pfs[1:]
        }

        for _, pf := range pfs {
                ports = pf.transform(ports)
        }

        sort.Sort(sort.IntSlice(ports))

        return ports, nil
}

func initports(startport, endport, webport int) []int {
        skip := map[int]struct{}{}
        // The main web port is a special case.
        skip[webport] = struct{}{}

        var ports []int
        for i := startport; i <= endport; i++ {
                if _, skipped := skip[i]; skipped {
                        continue
                }

                ports = append(ports, i)
        }

        return ports
}

const Portmax = 65535
const Portmin = 1
