package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/bgmerrell/tftpdmem/defs"
	errs "github.com/bgmerrell/tftpdmem/server/errors"
	"github.com/bgmerrell/tftpdmem/util"
)

type OpToHandleMap map[uint16]func(
	buf []byte, conn *net.UDPConn, src *net.UDPAddr) ([]byte, error)

type Server struct {
	port             int
	conn             *net.UDPConn
	opToHandle       OpToHandleMap
	isTransferServer bool
	StopCh           chan struct{}
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
			buf := make([]byte, defs.DatagramSize)
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
	op, err := readOpCode(buf)
	if err != nil || op < defs.MinOpCode || op > defs.MaxOpCode {
		s.respondWithErr(&errs.SrvError{defs.ErrIllegalOp, err.Error()}, src)
		return
	}
	fn, ok := s.opToHandle[op]
	if !ok {
		msg := fmt.Sprintf("Unsupported op: %d", op)
		log.Println(msg)
		s.respondWithErr(errors.New(msg), src)
		return
	}
	resp, err := fn(buf[defs.OpCodeSize:], s.conn, src)
	if err != nil {
		log.Println("Handle error: " + err.Error())
		s.respondWithErr(err, src)
		return
	}
	// No response if nil
	if resp == nil {
		// A transfer server returning nil means we're done (e.g.,
		// we just received a terminal ACK from client)
		if s.isTransferServer {
			s.StopCh <- struct{}{}
		}
	} else {
		err = s.respond(resp, src)
		if err != nil {
			log.Println(err)
			s.respondWithErr(err, src)
			return
		}
	}
	// We're done if we get an undersized data packet
	if (op == defs.OpData && len(buf) < defs.DatagramSize) && s.isTransferServer {
		s.StopCh <- struct{}{}
	}
}

func (s *Server) respond(resp []byte, src *net.UDPAddr) error {
	n, err := s.conn.WriteToUDP(resp, src)
	if err != nil || n != len(resp) {
		var msg string
		if err != nil {
			msg = "Error writing to UDP connection: " + err.Error()
		} else {
			msg = fmt.Sprintf(
				"Problem writing to UDP connection, %d of %d bytes written",
				n, len(resp))
		}
		log.Println(msg)
	}
	return err
}

func (s *Server) respondWithErr(err error, src *net.UDPAddr) {
	var srvErr *errs.SrvError
	shouldStop := s.isTransferServer
	switch err := err.(type) {
	case *errs.SrvError:
		srvErr = err
	case errs.UnexpectedRemoteTidErr:
		// Don't stop in the UnexpectedRemoteTidErr case
		shouldStop = false
		srvErr = &errs.SrvError{defs.ErrUnknownTid, err.Error()}
	default:
		srvErr = &errs.SrvError{defs.ErrGeneric, err.Error()}

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

func readOpCode(buf []byte) (op uint16, err error) {
	br := bytes.NewReader(buf)
	err = binary.Read(br, binary.BigEndian, &op)
	if err != nil {
		msg := "Unable to read op code: " + err.Error()
		log.Println(msg)
		return op, errors.New(msg)
	}
	return op, err
}
