package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		usage("expected sub-command")
	}
	switch os.Args[1] {
	case "receive":
		exit(receive(os.Args[2:]))
	case "send":
		exit(send(os.Args[2:]))
	default:
		usage(fmt.Sprintf("invalid sub-command: %q", os.Args[1]))
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
	fmt.Fprintf(os.Stderr, "# Listening: %q\n", operands[0])
	conn, err := l.Accept()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "# Accepted connection: %q\n", conn.RemoteAddr())
	buf := make([]byte, 32*1024)
	cr, err := io.CopyBuffer(os.Stdout, conn, buf)
	if cr > 0 {
		fmt.Fprintf(os.Stderr, "# received %d bytes\n", cr)
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
	fmt.Fprintf(os.Stderr, "# Connected: %q\n", conn.RemoteAddr())
	buf := make([]byte, 32*1024)
	cr, err := io.CopyBuffer(conn, os.Stdin, buf)
	if cr > 0 {
		fmt.Fprintf(os.Stderr, "# sent %d bytes\n", cr)
	}
	return err
}
