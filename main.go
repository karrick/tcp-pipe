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

func verbose(format string, a ...interface{}) {
	if *optVerbose {
		_, _ = fmt.Fprintf(os.Stderr, format, a...)
	}
}

func withGzipReader(use bool, ior io.Reader, callback func(ior io.Reader) error) error {
	if !use {
		return callback(ior)
	}
	verbose("# Using gzip compression\n")
	z, err := gzip.NewReader(ior)
	if err != nil {
		return err
	}
	err = callback(z)
	if err2 := z.Close(); err == nil {
		err = err2
	}
	return err
}

func withGzipWriter(use bool, iow io.Writer, callback func(iow io.Writer) error) error {
	if !use {
		return callback(iow)
	}
	verbose("# Using gzip compression\n")
	z := gzip.NewWriter(iow)
	err := callback(z)
	if err2 := z.Close(); err == nil {
		err = err2
	}
	return err
}

func withDial(remote string, callback func(iow io.Writer) error) error {
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		return err
	}
	verbose("# Connected: %q\n", conn.RemoteAddr())

	err = withGzipWriter(*optZip, conn, func(iow io.Writer) error {
		return callback(iow)
	})

	if err2 := conn.Close(); err == nil {
		err = err2
	}
	return err
}

func withListen(bind string, callback func(ior io.Reader) error) error {
	l, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}
	verbose("# Listening: %q\n", bind)
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	verbose("# Accepted connection: %q\n", conn.RemoteAddr())

	err = withGzipReader(*optZip, conn, func(ior io.Reader) error {
		return callback(ior)
	})

	if err2 := conn.Close(); err == nil {
		err = err2
	}
	return err
}

func receive(operands []string) error {
	if len(operands) < 1 {
		usage(fmt.Sprintf("cannot receive without binding address"))
	}
	return withListen(operands[0], func(ior io.Reader) error {
		cr, err := io.Copy(os.Stdout, ior)
		if cr > 0 {
			verbose("# Received %d bytes\n", cr)
		}
		return err
	})
}

func send(operands []string) error {
	if len(operands) < 1 {
		usage(fmt.Sprintf("cannot send without destination address"))
	}
	return withDial(operands[0], func(iow io.Writer) error {
		cr, err := io.Copy(iow, os.Stdin)
		if cr > 0 {
			verbose("# Sent %d bytes\n", cr)
		}
		return err
	})
}
