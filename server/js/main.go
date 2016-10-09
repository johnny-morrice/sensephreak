package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/johnny-morrice/sensephreak/scanner"
)

func main() {
	js.Global.Set("ScanBuilder", ScanBuilder)
}

type ScanOpts struct {
	scanner.Scan

	OnSuccess func(goodports, badports []int)
	OnError func(err string)
}

func ScanBuilder() *js.Object {
	return js.MakeWrapper(&ScanOpts{})
}

func (so *ScanOpts) SetHostname(hostname string) {
	so.Host = hostname
}

func (so *ScanOpts) SetConns(conns int) {
	so.Conns = conns
}

func (so *ScanOpts) SetApiport(apiport int) {
	so.Apiport = apiport
}

func (so *ScanOpts) SetStartPort(startPort int) {
	so.StartPort = startPort
}

func (so *ScanOpts) SetEndPort(endPort int) {
	so.EndPort = endPort
}

func (so *ScanOpts) SetOnSuccess(onSuccess func(goodports, badports []int)) {
	so.OnSuccess = onSuccess
}

func (so *ScanOpts) SetOnError(onError func(err string)) {
	so.OnError = onError
}

func (so *ScanOpts) WebScan() {
        go func() {
		var badports []int
		err := so.Launch()

		if err != nil {
			goto ERROR
		}

        	badports, err = so.Scanall()

        	if err == nil {
			goodports := so.GoodPorts(badports)

			so.OnSuccess(goodports, badports)

			return
        	}

ERROR:
		so.OnError(err.Error())
        }()
}
