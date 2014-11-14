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
	"github.com/bgmerrell/tftpdmem/server"
	errs "github.com/bgmerrell/tftpdmem/server/errors"
	"github.com/bgmerrell/tftpdmem/handlers/common"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func genTid() uint16 {
	return uint16(rand.Intn(1 << 16)) // RFC: "they must be between 0 and 65,535"
}

// startNewTransferServer starts a new server for the transferring data.  A
// reference to the new server is returned.
func startNewTransferServer(
		conn *net.UDPConn,
		handler func(buf []byte, conn *net.UDPConn, src *net.UDPAddr) error) *server.Server {
	// Set up transfer server that handles data requests
	opToHandle := server.OpToHandleMap{defs.OpData: handler}
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


func HandleWriteRequest(buf []byte, conn *net.UDPConn, src *net.UDPAddr) (err error) {
	n := bytes.Index(buf, []byte{0})
	if n < 1 {
		return &errs.SrvError{defs.ErrGeneric, "No filename provided"}
	}
	filename := string(buf[:n])
	buf = buf[n+1:]
	n = bytes.Index(buf, []byte{0})
	if n < 1 {
		return &errs.SrvError{defs.ErrGeneric, "No mode provided"}
	}
	mode := string(buf[:n])
	if mode != "octet" {
		return &errs.SrvError{defs.ErrGeneric,
			fmt.Sprintf("Unsupported mode: %s", mode)}
	}
	log.Printf("Write request for filename: %s, mode: %s", filename, mode)

	conn, err = initTransferConn(src)

	// Check to see if file already exists
	if fm.FileExists(filename) {
		return &errs.SrvError{defs.ErrFileExists,
			fmt.Sprintf("Filename \"%s\" already exists", filename)}
	}

	resp, err := common.BuildAckPacket(0)
	if err != nil {
		return err
	}

	s := startNewTransferServer(conn, HandleDataRequest)
	localPort := conn.LocalAddr().(*net.UDPAddr).Port

	// Add conn info to the file manager
	fm.AddConnInfo(localPort, src.Port, filename)

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
		return errors.New(msg)
	}
	return nil
}
