// Package common contains some common handler code
package common

import (
	"errors"
	"log"

	"github.com/bgmerrell/tftpdmem/defs"
	"github.com/bgmerrell/tftpdmem/util"
)

// BuildAckPacket builds and returns a TFTP ACK packet
func BuildAckPacket(blockNum uint16) ([]byte, error) {
	data := []interface{}{
		uint16(defs.OpAck),
		uint16(blockNum)}
	ackpkt, err := util.BuildResponse(data)
	if err != nil {
		msg := "Error building ack response: " + err.Error()
		log.Println(msg)
		return nil, errors.New(msg)
	}
	return ackpkt, err
}
