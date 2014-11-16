// The handlers packages contains various handlers for TFTP op codes.
package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/bgmerrell/tftpdmem/defs"
	fm "github.com/bgmerrell/tftpdmem/filemanager"
	"github.com/bgmerrell/tftpdmem/handlers/common"
	"github.com/bgmerrell/tftpdmem/server"
	errs "github.com/bgmerrell/tftpdmem/server/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func HandleWriteRequest(buf []byte, conn *net.UDPConn, src *net.UDPAddr) ([]byte, error) {
	return handleRequest(buf, conn, src, true)
}

func HandleReadRequest(buf []byte, conn *net.UDPConn, src *net.UDPAddr) ([]byte, error) {
	return handleRequest(buf, conn, src, false)
}

// startNewTransferServer starts a new server for the transferring data.  A
// reference to the new server is returned.
func startNewTransferServer(
	conn *net.UDPConn,
	opCode uint16,
	handler func(buf []byte, conn *net.UDPConn, src *net.UDPAddr) ([]byte, error)) *server.Server {
	// Set up transfer server that handles data requests
	opToHandle := server.OpToHandleMap{opCode: handler}
	localPort := conn.LocalAddr().(*net.UDPAddr).Port
	s := server.New(localPort, conn, opToHandle, true)
	go s.Serve()
	return s
}

// initTransferConn returns a new UDP conn to be used for data transfer.
func initTransferConn(src *net.UDPAddr) (*net.UDPConn, error) {
	// Get new conn on new port for transfer server
	laddr := net.UDPAddr{IP: net.ParseIP("127.0.0.1")}
	conn, err := net.ListenUDP(src.Network(), &laddr)
	if err != nil {
		msg := "Error getting new UDP conn: " + err.Error()
		log.Println(msg)
		return conn, errors.New(msg)
	}
	return conn, err
}

func handleRequest(buf []byte, conn *net.UDPConn, src *net.UDPAddr, isWrite bool) (resp []byte, err error) {
	n := bytes.Index(buf, []byte{0})
	if n < 1 {
		return nil, &errs.SrvError{defs.ErrGeneric, "No filename provided"}
	}
	filename := string(buf[:n])
	buf = buf[n+1:]
	n = bytes.Index(buf, []byte{0})
	if n < 1 {
		return nil, &errs.SrvError{defs.ErrGeneric, "No mode provided"}
	}
	mode := string(buf[:n])
	if mode != "octet" {
		return nil, &errs.SrvError{defs.ErrGeneric,
			fmt.Sprintf("Unsupported mode: %s", mode)}
	}

	if isWrite {
		log.Printf("Write request for filename: %s, mode: %s", filename, mode)
	} else {
		log.Printf("Read request for filename: %s, mode: %s", filename, mode)
	}

	conn, err = initTransferConn(src)
	localPort := conn.LocalAddr().(*net.UDPAddr).Port

	// Check if file exists
	exists := fm.FileExists(filename)
	if isWrite && exists {
		return nil, &errs.SrvError{defs.ErrFileExists,
			fmt.Sprintf("Filename \"%s\" already exists", filename)}
	} else if !isWrite && !exists {
		return nil, &errs.SrvError{defs.ErrFileNotFound,
			fmt.Sprintf("Filename \"%s\" does not exists", filename)}
	}

	// Add conn info to the file manager
	var nextBlockNum uint16
	if isWrite {
		nextBlockNum = 1
	} else {
		nextBlockNum = 0
	}
	fm.AddConnInfo(localPort, src.Port, filename, nextBlockNum)

	if isWrite {
		resp, err = common.BuildAckPacket(0)
		if err != nil {
			return nil, err
		}
	} else {
		data, err := fm.Read(localPort, src.Port, 0)
		if err != nil {
			return nil, err
		}
		resp, err = common.BuildDataPacket(defs.FirstDataBlock, data)
		if err != nil {
			return nil, err
		}
	}

	var s *server.Server
	if isWrite {
		s = startNewTransferServer(conn, defs.OpData, HandleWriteData)
	} else {
		s = startNewTransferServer(conn, defs.OpAck, HandleReadData)
	}

	n, err = conn.WriteToUDP(resp, src)
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
		fm.DelConnInfo(localPort)
		s.StopCh <- struct{}{}
		conn.Close()
		return nil, errors.New(msg)
	}
	return nil, nil
}
