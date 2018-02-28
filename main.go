package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/karrick/golf"
)

var (
	optHelp    = golf.BoolP('h', "help", false, "print help then exit")
	optVerbose = golf.BoolP('v', "verbose", false, "print verbose information")
	optZip     = golf.BoolP('z', "gzip", false, "(de-)compress with gzip")
)

func main() {
	golf.Parse()

	args := golf.Args()
	if len(args) == 0 {
		usage("expected sub-command")
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "receive":
		exit(receive(args))
	case "send":
		exit(send(args))
	default:
		usage(fmt.Sprintf("invalid sub-command: %q", cmd))
	}
}

func verbose(format string, a ...interface{}) {
	if *optVerbose {
		_, _ = fmt.Fprintf(os.Stderr, format, a...)
	}
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func usage(message string) {
	fmt.Fprintf(os.Stderr, "%s\nusage: %s [receive $binding_address | send $destination_address]\n", message, filepath.Base(os.Args[0]))
	os.Exit(2)
}

func receive(operands []string) error {
	if len(operands) < 1 {
		usage(fmt.Sprintf("cannot receive without binding address"))
	}
	l, err := net.Listen("tcp", operands[0])
	if err != nil {
		return err
	}
	verbose("# Listening: %q\n", operands[0])
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	verbose("# Accepted connection: %q\n", conn.RemoteAddr())
	var ior io.Reader = conn
	if *optZip {
		verbose("# Using gzip compression\n")
		ior, err = gzip.NewReader(ior)
		if err != nil {
			_ = conn.Close()
			return err
		}
	}
	cr, err := io.Copy(os.Stdout, ior)
	if cr > 0 {
		verbose("# Received %d bytes\n", cr)
	}
	if *optZip {
		if err2 := ior.(*gzip.Reader).Close(); err == nil {
			err = err2
		}
	}
	if err2 := conn.Close(); err == nil {
		err = err2
	}
	return err
}

func send(operands []string) error {
	if len(operands) < 1 {
		usage(fmt.Sprintf("cannot send without destination address"))
	}
	conn, err := net.Dial("tcp", operands[0])
	if err != nil {
		return err
	}
	verbose("# Connected: %q\n", conn.RemoteAddr())
	var iow io.Writer = conn
	if *optZip {
		verbose("# Using gzip compression\n")
		iow = gzip.NewWriter(iow)
	}
	cr, err := io.Copy(iow, os.Stdin)
	if cr > 0 {
		verbose("# Sent %d bytes\n", cr)
	}
	if *optZip {
		if err2 := iow.(*gzip.Writer).Close(); err == nil {
			err = err2
		}
	}
	if err2 := conn.Close(); err == nil {
		err = err2
	}
	return err
}
