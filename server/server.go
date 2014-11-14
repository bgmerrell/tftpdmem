package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/bgmerrell/tftpdmem/defs"
	fm "github.com/bgmerrell/tftpdmem/filemanager"
	"github.com/bgmerrell/tftpdmem/util"
)

const (
	opCodeBoundary = 2
)

type OpToHandleMap map[uint16]func(buf []byte, conn *net.UDPConn, src *net.UDPAddr) error

type Server struct {
	port             int
	conn             *net.UDPConn
	opToHandle       OpToHandleMap
	isTransferServer bool
	StopCh           chan struct{}
}

type SrvError struct {
	Code uint16
	Msg  string
}

func (e *SrvError) Error() string {
	return fmt.Sprintf("%s (%d)", e.Msg, e.Code)
}

func New(port int, conn *net.UDPConn, opToHandle OpToHandleMap, isTransferServer bool) *Server {
	return &Server{port,
		conn,
		opToHandle,
		isTransferServer,
		make(chan struct{})}
}

func (s *Server) Serve() {
	for {
		select {
		// Stop and close the connection
		case <-s.StopCh:
			s.StopCh <- struct{}{}
			s.conn.Close()
			return
		default:
			buf := make([]byte, defs.BlockSize)
			n, addr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				msg := "Error reading from UDP: " + err.Error()
				log.Println(msg)
				if s.isTransferServer {
					s.respondWithErr(errors.New(msg), addr)
					return
				}
				continue
			}
			go s.route(buf[:n], addr)
		}
	}
}

func (s *Server) route(buf []byte, src *net.UDPAddr) {
	var op uint16
	bufSize := len(buf)
	br := bytes.NewReader(buf)
	err := binary.Read(br, binary.BigEndian, &op)
	if err != nil {
		log.Println("Unable to read op code")
		s.respondWithErr(
			&SrvError{defs.ErrGeneric, err.Error()}, src)
		return
	}
	if op < defs.MinOpCode || op > defs.MaxOpCode {
		log.Println("Bad op code:", op)
		s.respondWithErr(&SrvError{defs.ErrIllegalOp, ""}, src)
		return
	}
	fn, ok := s.opToHandle[op]
	if !ok {
		msg := fmt.Sprintf("Unsupported op: %d", op)
		log.Println(msg)
		s.respondWithErr(errors.New(msg), src)
		return
	}
	err = fn(buf[opCodeBoundary:], s.conn, src)
	if err != nil {
		log.Println("Handle error: " + err.Error())
		s.respondWithErr(err, src)
		return
	}
	if s.isTransferServer && bufSize < defs.BlockSize {
		s.StopCh <- struct{}{}
	}
}

func (s *Server) respondWithErr(err error, src *net.UDPAddr) {
	var srvErr *SrvError
	shouldStop := s.isTransferServer
	switch err := err.(type) {
	case *SrvError:
		srvErr = err
	case fm.UnexpectedRemoteTidErr:
		// Don't stop in the UnexpectedRemoteTidErr case
		shouldStop = false
		srvErr = &SrvError{defs.ErrUnknownTid, err.Error()}
	default:
		srvErr = &SrvError{defs.ErrGeneric, err.Error()}

	}
	rawMsg := []byte(srvErr.Msg)
	data := []interface{}{
		uint16(defs.OpErr),
		srvErr.Code,
		rawMsg,
		uint8(0)}
	resp, err := util.BuildResponse(data)
	if err != nil {
		log.Printf("err building response: %s\n", err)
	}
	n, err := s.conn.WriteToUDP(resp, src)
	if n != len(resp) {
		log.Printf("Problem writing to UDP connection, %d of %d bytes written", n, len(resp))
	}
	if err != nil {
		log.Println("Error writing to UDP connection: " + err.Error())
	}
	if shouldStop {
		s.StopCh <- struct{}{}
	}
}
