package main

import (
        "bytes"
        "fmt"
        "github.com/johnny-morrice/sensephreak/scanner"
        "github.com/gopherjs/gopherjs/js"
)

func main() {
	js.Global.Set("Webscan", Webscan)
}

func Webscan(hostname string, apiport int, ports []int) {
        failed, err := scanner.Scanall(hostname, apiport, ports)


        if (err == nil) {
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
}

func update(messageHtml string) {
        doc := js.Global.Get("document")
        status := doc.Call("getElementById", "status")
        status.Set("innerHTML", messageHtml)
}
