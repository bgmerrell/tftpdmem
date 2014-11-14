package handlers

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"

	"github.com/bgmerrell/tftpdmem/defs"
	fm "github.com/bgmerrell/tftpdmem/filemanager"
	"github.com/bgmerrell/tftpdmem/util"
)

func HandleDataRequest(buf []byte, conn *net.UDPConn, src *net.UDPAddr) error {
	var blockNum uint16
	const blockNumBoundary = 2 // Two bytes of block num
	br := bytes.NewReader(buf)
	err := binary.Read(br, binary.BigEndian, &blockNum)
	if err != nil {
		return errors.New("Error parsing data request data: " + err.Error())
	}
	// The rest of the buffer is the file data
	buf = buf[blockNumBoundary:]

	localPort := conn.LocalAddr().(*net.UDPAddr).Port
	err = fm.Write(localPort, src.Port, blockNum, buf)
	if err != nil {
		log.Println(err)
		return err
	}

	// ACK packet (see spec)
	data := []interface{}{
		uint16(defs.OpAck),
		blockNum}
	resp, err := util.BuildResponse(data)
	if err != nil {
		msg := "Error building data ack response: " + err.Error()
		log.Println(msg)
		return errors.New(msg)
	}
	n, err := conn.WriteToUDP(resp, src)
	if n != len(resp) {
		log.Printf("Problem writing to UDP connection, %d of %d byte written", n, len(resp))
		return errors.New("Problem writing to UDP connection")
	}
	if err != nil {
		msg := "Error writing to UDP connection: " + err.Error()
		log.Println(msg)
		return errors.New(msg)
	}

	return err
}
