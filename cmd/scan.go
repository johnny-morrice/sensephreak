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
        "github.com/johnny-morrice/sensephreak/server"
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
                var badports []int
                var err error

		remote, err = cmd.PersistentFlags().GetString("remote")

                if err != nil {
                        panic(err)
                }

		scan := scanner.Scan{}
                scan.Host = remote
                scan.Apiport = server.Webport
                scan.Ports = defaultports

		err = scan.Launch()

                badports, err = scan.Scanall()

                if err != nil {
                        fmt.Fprintf(os.Stderr, "Error: %v", err)
                }

                for _, p := range badports {
                        fmt.Printf("%v\n", p)
                }
	},
}

func init() {
	RootCmd.AddCommand(scanCmd)

	scanCmd.PersistentFlags().String("remote", "localhost", "Remote host against which to scan")
}
