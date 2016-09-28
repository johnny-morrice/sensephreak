package main

import (
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/johnny-morrice/sensephreak/scanner"
)

func main() {
	js.Global.Set("Webscan", Webscan)
}

func Webscan(opts map[string]interface{}, success func(badports []int), failure func(err string)) {
	scan := &scanner.Scan{}

	var badparam string

	params := make([]interface{}, 4)
	keys := []string {"hostname", "apiport", "conns", "ports"}

	var ok bool
	var host string
	var apiport int
	var conns int
	var ports []interface{}

	for i, k := range keys {
		p, ok := opts[k]

		if !ok {
			badparam = k
			break
		}

		params[i] = p
	}

	if badparam != "" {
		failure(fmt.Sprintf("Missing parameter: %v", badparam))
		return
	}

	host, ok = params[0].(string)

	if !ok {
		badparam = "hostname"
		goto BADINPUT
	}

	apiport, ok = params[1].(int)

	if !ok {
		badparam = "apiport"
		goto BADINPUT
	}

	conns, ok = params[2].(int)

	if !ok {
		badparam = "conns"
		goto BADINPUT
	}

	ports, ok = params[3].([]interface{})

	if !ok {
		badparam = "ports"
		goto BADINPUT
	}

	BADINPUT:
	if badparam != "" {
		failure(fmt.Sprintf("Invalid parameter: %v", badparam))
		return
	}

	scan.Host = host
	scan.Conns = conns
	scan.Apiport = apiport

	scan.Ports = make([]int, len(ports))

	for i, any := range ports {
		port, ok := any.(int)

		if !ok {
			failure(fmt.Sprintf("Bad port: %v", any))
			return
		}

		scan.Ports[i] = port
	}

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
