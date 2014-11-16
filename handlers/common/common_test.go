package common

import (
	"bytes"
	"testing"
)

func TestBuildAckPacket(t *testing.T) {
	blockNum := uint16(123)
	ackpkt, err := BuildAckPacket(blockNum)
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0x00, 0x04, 0x00, 0x7b}
	if bytes.Compare(ackpkt, expected) != 0 {
		t.Errorf("Got %#v, want %#v", ackpkt, expected)
	}
}

func TestBuildDataPacket(t *testing.T) {
	blockNum := uint16(123)
	data := []byte("abc")
	datapkt, err := BuildDataPacket(blockNum, data)
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0x00, 0x03, 0x00, 0x7b, 0x61, 0x62, 0x63}
	if bytes.Compare(datapkt, expected) != 0 {
		t.Errorf("Got %#v, want %#v", datapkt, expected)
	}
}
