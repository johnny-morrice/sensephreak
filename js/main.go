package main

import (
	"bytes"
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/johnny-morrice/sensephreak/scanner"
)

func main() {
	js.Global.Set("Webscan", Webscan)
}

func Webscan(opts *js.Object) {
	scan := &scanner.Scan{}

	scan.Host = opts.Get("hostname").String()
	scan.Apiport = opts.Get("apiport").Int()
	scan.Conns = opts.Get("conns").Int()

	portopts := opts.Get("ports")
	scan.Ports = make([]int, portopts.Length())

	for i := 0; i < portopts.Length(); i++ {
		scan.Ports[i] = portopts.Index(i).Int()
	}

	update("Scanning...")

	scan.Launch()

        go func() {
        	failed, err := scan.Scanall()

        	if err == nil {
        		if len(failed) == 0 {
        			update("All tested ports are free.")
        		} else {
        			buff := &bytes.Buffer{}
        			buff.WriteString("Failed ports:<br/>")

        			for _, p := range failed {
        				fmt.Fprintf(buff, "%v<br/>", p)
        			}

        			update(buff.String())
        		}
        	} else {
        		update("There was an error")
        		panic(err)
        	}
        }()
}

func update(messageHtml string) {
	doc := js.Global.Get("document")
	status := doc.Call("getElementById", "status")
	status.Set("innerHTML", messageHtml)
}
