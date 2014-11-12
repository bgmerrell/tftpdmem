package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/bgmerrell/tftpdmem/codes"
)

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

type srvError struct {
	code uint16
	msg  string
}

func (e *srvError) Error() string {
	return fmt.Sprintf("%s (%d)", e.msg, e.code)
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
		s.respondWithErr(&srvError{codes.ErrGeneric, "Server too busy"})
		return
	}
	var op uint16
	br := bytes.NewReader(buf)
	err := binary.Read(br, binary.BigEndian, &op)
	if err != nil {
		log.Println("Unable to read op code")
		s.respondWithErr(&srvError{codes.ErrGeneric, err.Error()})
		return
	}
	if op < minOpCode || op > maxOpCode {
		log.Println("Bad op code:", op)
		s.respondWithErr(&srvError{codes.ErrIllegalOp, ""})
		return
	}
	s.handleOpCode(op, buf[opCodeBoundary:], src)
}

func write(filename string, data []byte) error {
	log.Println("Writing filename:", filename)
	log.Println("Writing data:", data)
	return nil
}

func (s *Server) handleOpCode(op uint16, buf []byte, src *net.UDPAddr) {
	var err error
	switch op {
	case codes.OpRrq:
		log.Println("Read request")
	case codes.OpWrq:
		err = handleWriteRequest(s.conn, buf, src)
	case codes.OpData:
		err = handleDataRequest(s.conn, buf, src)
		log.Println("Data msg")
	case codes.OpAck:
		log.Println("Ack msg")
	case codes.OpErr:
		log.Println("Err msg")
	}
	if err != nil {
		log.Printf("%#v\n", err)
		s.respondWithErr(err)
	}
}

func (s *Server) respondWithErr(err error) {
	var srvErr *srvError
	switch err.(type) {
	case *srvError:
		srvErr = err.(*srvError)
	default:
		srvErr = &srvError{codes.ErrGeneric, err.Error()}

	}
	rawMsg := []byte(srvErr.msg)
	data := []interface{}{
		uint16(codes.OpErr),
		srvErr.code,
		rawMsg,
		uint8(0)}
	resp, err := buildResponse(data)
	if err != nil {
		log.Printf("err building response: %s\n", err)
	}
	// TODO: respond
	log.Printf("err response: %#v\n", resp)

}

func buildResponse(data []interface{}) ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		if err != nil {
			break
		}
	}
	return buf.Bytes(), err
}
