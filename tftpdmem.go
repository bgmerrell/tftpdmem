package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/bgmerrell/tftpdmem/server"
)

// flags
var (
	port            int
	maxConcurrent   uint
	responseTimeout uint
)

func init() {
	flag.IntVar(&port, "port", 69, "Port for the tftp server")
	flag.UintVar(&maxConcurrent, "maxConcurrent", 128, "Max requests handled concurrently")
	flag.UintVar(&responseTimeout, "responseTimeout", 10, "Number of seconds to allow a request wait before being handled")
	flag.Parse()
}

func main() {
	log.Println("Starting tftpdmem on port", port)
	laddr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port}
	conn, err := net.ListenUDP(laddr.Network(), &laddr)
	defer conn.Close()
	if err != nil {
		log.Println("ListenUDP failure", err)
		conn.Close()
		os.Exit(1)
	}

	s := server.New(port, maxConcurrent, responseTimeout, conn)
	quitCh := make(chan struct{})
	go s.Serve(quitCh)

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	log.Println(<-sigCh)
}
