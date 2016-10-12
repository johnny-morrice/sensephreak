package util

type LaunchData struct {
	StartPort int
	EndPort int
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
