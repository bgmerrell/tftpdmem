package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/bgmerrell/tftpdmem/defs"
	"github.com/bgmerrell/tftpdmem/handlers"
	"github.com/bgmerrell/tftpdmem/server"
)

// flags
var (
	port int
)

func init() {
	flag.IntVar(&port, "port", 69, "Port for the tftp server")
	flag.Parse()
}

func main() {
	log.Println("Starting tftpdmem on port", port)
	laddr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port}
	conn, err := net.ListenUDP(laddr.Network(), &laddr)
	if err != nil {
		log.Println("ListenUDP failure", err)
		os.Exit(1)
	}
	defer conn.Close()

	// The "main" server only supports ACK and read and write requests,
	// which create new servers for data transfer.
	opToHandle := server.OpToHandleMap{
		defs.OpWrq: handlers.HandleWriteRequest,
		defs.OpRrq: handlers.HandleReadRequest,
		// We'll just ignore ACKs to the main server, this server isn't
		// smart enough to do anything about them.
		defs.OpAck: func([]byte, *net.UDPConn, *net.UDPAddr) ([]byte, error) { return nil, nil }}
	s := server.New(port, conn, opToHandle, false)
	go s.Serve()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	log.Println(<-sigCh)
}
