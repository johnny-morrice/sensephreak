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
        "crypto/rand"
	"fmt"
        "math/big"
        "net"
	"os"

        "github.com/johnny-morrice/sensephreak/util"
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
                serveargs, err := getserveargs(cmd)

                if err != nil {
                        fmt.Fprintf(os.Stderr, "Error processing arguments\n", err)

			return
                }

                launchserver(serveargs)
	},
}

func launchserver(args *serveparams) {
        if args.secret == randomsecret {
                args.secret = cryptrandstr(20)
        }

        ports, err := util.Ports(args.ports, int(args.webport))

        if err != nil {
                fmt.Fprintf(os.Stderr, "Error in port specification: %v\n", err)

                return
        }

        s := server.Server{}
        s.Bind = args.bind
        s.Hostname = args.hostname
        s.Webport = int(args.webport)
        s.Ports = ports
        s.Secret = args.secret
        s.Title = args.title
        s.Heading = args.heading
	s.UseTLS = args.tls
	s.Certfile = args.certfile
	s.Keyfile = args.keyfile

        s.Serve()
}

func getserveargs(cmd *cobra.Command) (*serveparams, error) {
        var err error
        persistent := cmd.PersistentFlags()

        args := &serveparams{}

        args.bind, err = persistent.GetIP("bind")

        if err != nil {
                return nil, err
        }
        args.webport, err = persistent.GetUint("webport")

        if err != nil {
                return nil, err
        }

        args.hostname, err = persistent.GetString("hostname")

        if err != nil {
                return nil, err
        }

        args.secret, err = persistent.GetString("secret")

        if err != nil {
                return nil, err
        }

        args.title, err = persistent.GetString("title")

        if err != nil {
                return nil, err
        }

        args.heading, err = persistent.GetString("heading")

        if err != nil {
                return nil, err
        }

        args.ports, err = persistent.GetString("ports")

	if err != nil {
		return nil, err
	}

	args.tls, err = persistent.GetBool("tls")

	if err != nil {
		return nil, err
	}

	args.certfile, err = persistent.GetString("certfile")

	if err != nil {
		return nil, err
	}

	args.keyfile, err = persistent.GetString("keyfile")

        return args, err
}

type serveparams struct {
        bind net.IP
        webport uint
        hostname string
        secret string
        title string
        heading string
        ports string
	tls bool
	certfile string
	keyfile string
}

func init() {
	RootCmd.AddCommand(serveCmd)

	persistent := serveCmd.PersistentFlags()
	persistent.IP("bind", net.IP([]byte{127, 0, 0, 1}), "Interface on which to listen")
        persistent.String("hostname", "localhost", "External hostname (Mandatory for CORS)")
	persistent.Uint("webport", 80, "Web port")
        persistent.String("secret", randomsecret, "Cookie cache secret")
        persistent.String("title", "Outgoing Port Block Scanner", "Index page title")
        persistent.String("heading", "Sensesphreak: single-exe outgoing port block scanner", "Index page heading")
        persistent.String("ports", "+1:65535", "Ports.  Format: +Port[:Range],-Port[:Range]... Starting with + overrides defaults.")
	persistent.Bool("tls", false, "Use HTTPS")
	persistent.String("certfile", "cert.pem", "HTTPS certificate")
	persistent.String("keyfile", "key.pem", "HTTPS key")
}

// Strongly random digit (0-9) string of the given length.
func cryptrandstr(length int) string {
        var out string

        max := big.NewInt(9)
        for i := 0; i < length; i++ {
                next, err := rand.Int(rand.Reader, max)

                if err != nil {
                        panic(err)
                }

                nextstr := next.String()

                out = out + nextstr
        }

        return out
}

const randomsecret = "(random)"
