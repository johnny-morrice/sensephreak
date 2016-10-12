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
        "github.com/johnny-morrice/sensephreak/util"
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
		scanargs, err := getscanargs(cmd)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing arguments: %v\n", err)

			return
		}

		err = launchscan(scanargs)

                if err != nil {
                        fmt.Fprintln(os.Stderr, err)
                }
	},
}

func launchscan(args *scanparam) error {
        scan := &scanner.Scan{}

        scan.Host = args.remote
        scan.Apiport = int(args.webport)
        scan.StartPort = int(args.startport)
        scan.EndPort = int(args.endport)
        scan.Conns = int(args.conns)
        scan.Verbose = args.verbose

        err := scan.Launch()

	listports, err := scan.Scanall()

	if err != nil {
		return err
	}

	if args.good {
		listports = scan.GoodPorts(listports)
	}

	for _, p := range listports {
		fmt.Printf("%v\n", p)
	}

	return nil
}

func getscanargs(cmd *cobra.Command) (*scanparam, error) {
	persistent := cmd.PersistentFlags()

	args := &scanparam{}

        var err error
	args.remote, err = persistent.GetString("remote")

	if err != nil {
		return nil, err
	}

	args.good, err = persistent.GetBool("good")

	if err != nil {
		return nil, err
	}

	args.verbose, err = persistent.GetBool("verbose")

	if err != nil {
		return nil, err
	}

	args.startport, err = persistent.GetUint("startport")

	if err != nil {
		return nil, err
	}

	args.endport, err = persistent.GetUint("endport")

	if err != nil {
		return nil, err
	}

	args.conns, err = persistent.GetUint("conns")

	if err != nil {
		return nil, err
	}

	args.webport, err = persistent.GetUint("webport")

	return args, err
}

type scanparam struct {
	remote string
	good bool
	verbose bool
	startport uint
	endport uint
	conns uint
	webport uint
}

func init() {
	RootCmd.AddCommand(scanCmd)

	persistent := scanCmd.PersistentFlags()

	persistent.String("remote", "localhost", "Remote host against which to scan")
	persistent.Bool("good", false, "List ports that are not blocked.")
	persistent.Bool("verbose", false, "More information on the program operation.")
	persistent.Uint("startport", util.Portmin, "Start port")
	persistent.Uint("endport", util.Portmax, "End port")
	persistent.Uint("conns", scanner.DefaultConns, "Number of connections")
	persistent.Uint("webport", 80, "Web API port")

}
