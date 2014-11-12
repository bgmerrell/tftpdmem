package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/bgmerrell/tftpdmem/codes"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func genTid() uint16 {
	return uint16(rand.Intn(1 << 16)) // RFC: "they must be between 0 and 65,535"
}

func handleWriteRequest(conn *net.UDPConn, buf []byte, src *net.UDPAddr) error {
	n := bytes.Index(buf, []byte{0})
	if n < 1 {
		return &srvError{codes.ErrGeneric, "No filename provided"}
	}
	filename := string(buf[:n])
	buf = buf[n+1:]
	n = bytes.Index(buf, []byte{0})
	if n < 1 {
		return &srvError{codes.ErrGeneric, "No mode provided"}
	}
	mode := string(buf[:n])
	if mode != "octet" {
		return &srvError{codes.ErrGeneric,
			fmt.Sprintf("Unsupported mode: %s", mode)}
	}
	log.Printf("Write request for filename: %s, mode: %s", filename, mode)
	// ACK packet (see spec)
	data := []interface{}{
		uint16(codes.OpAck),
		uint16(0)}
	resp, err := buildResponse(data)
	if err != nil {
		msg := "Error building wrq ack response: " + err.Error()
		log.Println(msg)
		return errors.New(msg)
	}
	n, err = conn.WriteToUDP(resp, src)
	if n != len(resp) {
		log.Printf("Problem writing to UDP connection, %d of %d byte written", n, len(resp))
		return errors.New("Problem writing to UDP connection")
	}
	if err != nil {
		msg := "Error writing to UDP connection: " + err.Error()
		log.Println(msg)
		return errors.New(msg)
	}
	return nil
}

func handleDataRequest(conn *net.UDPConn, buf []byte, src *net.UDPAddr) error {
	var blockNum uint16
	const blockNumBoundary = 2 // Two bytes of block num
	br := bytes.NewReader(buf)
	err := binary.Read(br, binary.BigEndian, &blockNum)
	if err != nil {
		return errors.New("Error parsing data request data: " + err.Error())
	}
	// The rest of the buffer is the file data
	buf = buf[blockNumBoundary:]
	log.Printf("block num: %d\n", blockNum)
	log.Printf("data: %#v\n", buf)

	// TODO: do something with the data, make sure block num is in order

	// ACK packet (see spec)
	data := []interface{}{
		uint16(codes.OpAck),
		blockNum}
	resp, err := buildResponse(data)
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
