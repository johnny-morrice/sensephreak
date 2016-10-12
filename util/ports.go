package util

import (
        "fmt"
        "io"
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

func (pf portformat) transform(ports []int, webport int) []int {
        if pf.cmd == addport {
                for p := pf.startport; p <= pf.endport; p++ {
                        if p == webport {
                                continue
                        }

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
                                if p == webport {
                                        continue
                                }

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
                ports = pf.transform(ports, webport)
        }

        sort.Sort(sort.IntSlice(ports))

        return ports, nil
}

func initports(startport, endport, webport int) []int {
        var ports []int
        for port := startport; port <= endport; port++ {
                if webport == port {
                        continue
                }

                ports = append(ports, port)
        }

        return ports
}

type Portstate uint8

const (
        PortOk = Portstate(iota)
        PortBlocked
        PortOmitted
)

type PortStatus struct {
        Port int
        State Portstate
}

func (ps PortStatus) Write(w io.Writer) {
        var msg string
        switch (ps.State) {
        case PortOk:
                msg = "ok"
        case PortBlocked:
                msg = "block"
        case PortOmitted:
                msg = "omit"
        default:
                panic(fmt.Sprintf("Unknown Portstate: %v", ps.State))
        }

        fmt.Fprintf(w, "%v\t%v", ps.Port, msg)
}

func GoodPorts(portinfo []PortStatus) []PortStatus {
	var goodports []PortStatus

	for _, info := range portinfo {
		if info.State == PortOk {
			goodports = append(goodports, info)
		}
	}

	return goodports
}

func BadPorts(portinfo []PortStatus) []PortStatus {
	var badports []PortStatus

	for _, info := range portinfo {
		if info.State == PortBlocked || info.State == PortOmitted {
			badports = append(badports, info)
		}
	}

	return badports
}

const Portmax = 65535
const Portmin = 1
