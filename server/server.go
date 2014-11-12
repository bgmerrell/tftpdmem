package server

import (
	"bytes"
	"encoding/binary"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/bgmerrell/tftpdmem/codes"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	initBufSize    = 512
	minOpCode      = 1
	maxOpCode      = 5
	opCodeBoundary = 2
)

type Server struct {
	port        int
	connLimitCh chan struct{}
	respTimeout uint
	conn        *net.UDPConn
}

func New(port int, maxConcurrent uint, respTimeout uint, conn *net.UDPConn) *Server {
	return &Server{port, make(chan struct{}, maxConcurrent), respTimeout, conn}
}

func (s *Server) Serve(quitCh chan struct{}) {
	for {
		select {
		// We're done.  Stop reading from the UDP connection, it will
		// be closing soon.
		case <-quitCh:
			quitCh <- struct{}{}
			return
		default:
			buf := make([]byte, initBufSize)
			n, addr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				log.Println("Error reading from UDP:", err)
				continue
			}
			go s.handle(buf[:n], addr)
		}
	}
}

func (s *Server) handle(buf []byte, src *net.UDPAddr) {
	select {
	case s.connLimitCh <- struct{}{}:
		defer func() {
			<-s.connLimitCh
		}()
	case <-time.After(time.Duration(s.respTimeout) * time.Second):
		log.Println("Server too busy")
		s.respondWithErr(codes.ErrGeneric, "Server too busy")
		return
	}
	op, n := binary.Uvarint(buf[:opCodeBoundary])
	if n <= 0 || op < minOpCode || op > maxOpCode {
		log.Println("Invalid op code:", op)
		s.respondWithErr(codes.ErrIllegalOp, "")
		return
	}
	log.Println("bytes received:", len(buf))
	log.Println("op:", op)
	log.Printf("addr: %#v\n", src)
	log.Printf("data: %#v\n", buf)
	log.Println("----------------------------")
	s.handleOpCode(op, buf[opCodeBoundary:], src)
}

func (s *Server) handleOpCode(op uint64, buf []byte, src *net.UDPAddr) {
	switch op {
	case codes.OpRrq:
		log.Println("Read request")
	case codes.OpWrq:
		log.Println("Write request")
	case codes.OpData:
		log.Println("Data msg")
	case codes.OpAck:
		log.Println("Ack msg")
	case codes.OpErr:
		log.Println("Err msg")
	}
}

func (s *Server) respondWithErr(code uint16, msg string) {
	buf := &bytes.Buffer{}
	rawMsg := []byte(msg)
	// RFC: "terminated with a zero byte"
	rawMsg = append(rawMsg, 0)
	data := []interface{}{
		code,
		rawMsg,
		uint16(0)}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		if err != nil {
			log.Println("Failed to binary write:", err)
		}
	}
	log.Printf("err buf: %#v\n", buf)
}

func genTid() int {
	return rand.Intn(1 << 16) // RFC: "they must be between 0 and 65,535"
}
