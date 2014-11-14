package handlers

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"

	fm "github.com/bgmerrell/tftpdmem/filemanager"
	"github.com/bgmerrell/tftpdmem/handlers/common"
)

func getBlockNum(buf []byte) (uint16, error) {
	var blockNum uint16
	br := bytes.NewReader(buf)
	err := binary.Read(br, binary.BigEndian, &blockNum)
	if err != nil {
		return blockNum, errors.New(
			"Error parsing data request data: " + err.Error())
	}
	return blockNum, err
}

func HandleWriteData(buf []byte, conn *net.UDPConn, src *net.UDPAddr) error {
	blockNum, err := getBlockNum(buf)
	if err != nil {
		return err
	}
	// The rest of the buffer is the file data
	const blockNumBoundary = 2 // Two bytes of block num
	buf = buf[blockNumBoundary:]

	localPort := conn.LocalAddr().(*net.UDPAddr).Port
	err = fm.Write(localPort, src.Port, blockNum, buf)
	if err != nil {
		return err
	}

	resp, err := common.BuildAckPacket(blockNum)
	if err != nil {
		return err
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

func HandleReadData(buf []byte, conn *net.UDPConn, src *net.UDPAddr) error {
	blockNum, err := getBlockNum(buf)
	if err != nil {
		return err
	}

	localPort := conn.LocalAddr().(*net.UDPAddr).Port

	data, err := fm.Read(localPort, src.Port, blockNum)
	if err != nil {
		return err
	}
	// No response for a terminal ACK
	if data == nil {
		return nil
	}

	resp, err := common.BuildDataPacket(blockNum+1, data)
	if err != nil {
		return err
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
