// Copyright © 2016 John Morrice <john@functorama.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
        "os"

	"github.com/spf13/cobra"
        "github.com/johnny-morrice/sensephreak/scanner"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Command line utility to check blocked ports",
	Long: `Command line utility that scans against a sensephreak server
instance:

# Scan against remote sensephreak instance (assuming your firewall blocks
# outgoing connections to ports 1, 25, and 81)
$ sensephreak scan --remote yoursite.com
1
25
81`,
	Run: func(cmd *cobra.Command, args []string) {
                var remote string
		var good bool
		var startport uint
		var endport uint
		var conns uint
		var webport uint
                var listports []int
                var err error
		var scan *scanner.Scan

		persistent := cmd.PersistentFlags()

		remote, err = persistent.GetString("remote")
		good, err = persistent.GetBool("good")
		startport, err = persistent.GetUint("startport")
		endport, err = persistent.GetUint("endport")
		conns, err = persistent.GetUint("conns")
		webport, err = persistent.GetUint("webport")

                if err != nil {
                        goto ERROR
                }

                scan.Host = remote
                scan.Apiport = int(webport)
		scan.StartPort = int(startport)
		scan.EndPort = int(endport)
		scan.Conns = int(conns)

		err = scan.Launch()

                if err != nil {
                        goto ERROR
                }

                listports, err = scan.Scanall()

		if good {
			const start = 1
			const end = 65535
			listports = scan.GoodPorts(listports)
		}

		for _, p := range listports {
			fmt.Printf("%v\n", p)
		}
ERROR:
                if err != nil {
                        fmt.Fprintln(os.Stderr, err)
                }
	},
}

func init() {
	RootCmd.AddCommand(scanCmd)

	persistent := scanCmd.PersistentFlags()

	persistent.String("remote", "localhost", "Remote host against which to scan")
	persistent.Bool("good", false, "List ports that are not blocked.")
	persistent.Uint("startport", portmin, "Start port")
	persistent.Uint("endport", portmax, "End port")
	persistent.Uint("conns", scanner.DefaultConns, "Number of connections")
	persistent.Uint("webport", 80, "Web API port")

}
