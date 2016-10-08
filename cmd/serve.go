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
        "net"

        "github.com/johnny-morrice/sensephreak/server"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a sensephreak server instance.",
	Long: `Run a sensephreak server.  The server typically runs on a range of
(or all available) ports.  For this reason, it is easiest to run sensphreak
through docker:

# Default runs on localhost for safety (although this is unlikely to be useful
# in production.)
$ sensephreak serve

# Listen on all ports.
$ sensephreak serve --bind 0.0.0.0

# Use docker to listen on all ports in containerized application (in root of source code):
$ docker build -t sensephreak .
$ docker run --cap-add SYS_RESOURCE --name test --rm sensephreak serve --bind 0.0.0.0`,
	Run: func(cmd *cobra.Command, args []string) {
                var bind net.IP
                var hostname string
                var err error

                bind, err = cmd.PersistentFlags().GetIP("bind")

                if err != nil {
                        panic(err)
                }

                hostname, err = cmd.PersistentFlags().GetString("hostname")

                if err != nil {
                        panic(err)
                }

                server.Serve(bind, hostname, defaultports)
	},
}

var defaultports []int

func init() {
	RootCmd.AddCommand(serveCmd)

	persistent := serveCmd.PersistentFlags()
	persistent.IP("bind", net.IP([]byte{127, 0, 0, 1}), "Interface on which to listen")
        persistent.String("hostname", "localhost", "External hostname (Mandatory for CORS)")

        // Generate the ports at this point, even though these are
        // currently non-configurable.
        skip := map[int]struct{}{}
        // The main web port is a special case.
        skip[server.Webport] = struct{}{}

        for i := portmin; i < portmax; i++ {
                if _, skipped := skip[i]; skipped {
                        continue
                }

                defaultports = append(defaultports, i)
        }
}



const portmax = 65536
const portmin = 1
