package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/johnny-morrice/sensephreak/scanner"
)

func main() {
	js.Global.Set("Webscan", Webscan)
	js.Global.Set("GoodPorts", scanner.GoodPorts)
}

func Webscan(hostname string, conns, apiport int, ports []int, success func(badports []int), failure func(err string)) {
	scan := &scanner.Scan{}

	scan.Host = hostname
	scan.Apiport = apiport
	scan.Ports = ports
	scan.Conns = conns

        go func() {
		var badports []int
		err := scan.Launch()

		if err != nil {
			goto ERROR
		}

        	badports, err = scan.Scanall()

        	if err == nil {
			success(badports)
			return
        	}

ERROR:
		failure(err.Error())
        }()
}
