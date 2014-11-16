package handlers

import (
	"bytes"
	"encoding/binary"
	"errors"
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

func HandleWriteData(buf []byte, conn *net.UDPConn, src *net.UDPAddr) (resp []byte, err error) {
	blockNum, err := getBlockNum(buf)
	if err != nil {
		return nil, err
	}
	// The rest of the buffer is the file data
	const blockNumBoundary = 2 // Two bytes of block num
	buf = buf[blockNumBoundary:]

	localPort := conn.LocalAddr().(*net.UDPAddr).Port
	err = fm.Write(localPort, src.Port, blockNum, buf)
	if err != nil {
		return nil, err
	}

	resp, err = common.BuildAckPacket(blockNum)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func HandleReadData(buf []byte, conn *net.UDPConn, src *net.UDPAddr) (resp []byte, err error) {
	blockNum, err := getBlockNum(buf)
	if err != nil {
		return nil, err
	}

	localPort := conn.LocalAddr().(*net.UDPAddr).Port

	data, err := fm.Read(localPort, src.Port, blockNum)
	if err != nil {
		return nil, err
	}
	// No response for a terminal ACK
	if data == nil {
		return nil, nil
	}

	resp, err = common.BuildDataPacket(blockNum+1, data)
	if err != nil {
		return nil, err
	}

	return resp, err
}
