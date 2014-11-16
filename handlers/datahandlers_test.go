package handlers

import (
	"bytes"
	"fmt"
	"net"
	"testing"

	fmgr "github.com/bgmerrell/tftpdmem/filemanager"
	errs "github.com/bgmerrell/tftpdmem/server/errors"
)

func TestGetBlockNum(t *testing.T) {
	expectedBlockNum := uint16(0xa8ca)
	blockNum, err := getBlockNum([]byte{0xa8, 0xca, 0x61, 0x62, 0x3})
	if err != nil {
		t.Fatal(err)
	}
	if blockNum != expectedBlockNum {
		t.Errorf("block num: %d, want %d", blockNum, expectedBlockNum)
	}
}

func TestGetBlockNumFail(t *testing.T) {
	// Not enough bytes should fail
	_, err := getBlockNum([]byte{0xa8})
	if err == nil {
		t.Error("Expected unexpected EOF")
	}
}

func TestHandleWriteData(t *testing.T) {
	fm := fmgr.New()
	filename := "foo"
	nextBlockNum := uint16(1)
	ip := "127.0.0.1"
	data := []byte{0x00, 0x01, 0x61, 0x62, 0x63}
	// Expect an ack with block number set to 0
	expectedResp := []byte{0x00, 0x04, 0x00, 0x01}
	lAddr := &net.UDPAddr{IP: net.ParseIP(ip)}
	conn, err := net.ListenUDP(lAddr.Network(), lAddr)
	if err != nil {
		t.Fatal("Error getting new UDP conn:", err)
	}
	defer conn.Close()
	lAddr = conn.LocalAddr().(*net.UDPAddr)
	rAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: lAddr.Port - 1}
	err = fm.AddConnInfo(lAddr.Port, rAddr.Port, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := HandleWriteData(data, conn, rAddr, fm)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(resp, expectedResp) != 0 {
		t.Errorf("resp: %#v, want %#v", resp, expectedResp)
	}
}

func TestHandleWriteDataBadBlock(t *testing.T) {
	fm := fmgr.New()
	filename := "foo"
	nextBlockNum := uint16(1)
	ip := "127.0.0.1"
	data := []byte{0xa1}
	lAddr := &net.UDPAddr{IP: net.ParseIP(ip)}
	conn, err := net.ListenUDP(lAddr.Network(), lAddr)
	if err != nil {
		t.Fatal("Error getting new UDP conn:", err)
	}
	defer conn.Close()
	lAddr = conn.LocalAddr().(*net.UDPAddr)
	rAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: lAddr.Port - 1}
	err = fm.AddConnInfo(lAddr.Port, rAddr.Port, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	_, err = HandleWriteData(data, conn, rAddr, fm)
	if err == nil {
		t.Errorf("Expected error getting block for data: %#v", data)
	}
}

func TestHandleWriteDataUnexpectedRemoteTid(t *testing.T) {
	fm := fmgr.New()
	filename := "foo"
	nextBlockNum := uint16(1)
	ip := "127.0.0.1"
	data := []byte{0x00, 0x01, 0x61, 0x62, 0x63}
	lAddr := &net.UDPAddr{IP: net.ParseIP(ip)}
	conn, err := net.ListenUDP(lAddr.Network(), lAddr)
	if err != nil {
		t.Fatal("Error getting new UDP conn:", err)
	}
	defer conn.Close()
	lAddr = conn.LocalAddr().(*net.UDPAddr)
	rAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: lAddr.Port - 1}
	fmt.Println("ltid:", lAddr.Port)
	fmt.Println("rtid:", rAddr.Port)
	err = fm.AddConnInfo(lAddr.Port, rAddr.Port, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	// Set wrong remote TID
	rAddr = &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: lAddr.Port - 2}
	_, err = HandleWriteData(data, conn, rAddr, fm)
	if err == nil {
		t.Error("Expected error writing data with incorrect remote TID")
	}
	if _, ok := err.(errs.UnexpectedRemoteTidErr); !ok {
		t.Error("Expected UnexpectedRemoteTidErr writing data with incorrect TID")
	}
}

func TestHandleReadData(t *testing.T) {
	filename := "foo"
	fileBytes := []byte{0x61, 0x62, 0x63}
	fm := fmgr.NewWithExistingFiles(map[string][]byte{filename: fileBytes})
	nextBlockNum := uint16(0)
	ip := "127.0.0.1"
	data := []byte{0x00, 0x00}
	// Expect a data packet with block number set to 1 and some data
	expectedResp := []byte{0x00, 0x03, 0x00, 0x01, 0x61, 0x62, 0x63}
	lAddr := &net.UDPAddr{IP: net.ParseIP(ip)}
	conn, err := net.ListenUDP(lAddr.Network(), lAddr)
	if err != nil {
		t.Fatal("Error getting new UDP conn:", err)
	}
	defer conn.Close()
	lAddr = conn.LocalAddr().(*net.UDPAddr)
	rAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: lAddr.Port - 1}
	err = fm.AddConnInfo(lAddr.Port, rAddr.Port, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := HandleReadData(data, conn, rAddr, fm)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(resp, expectedResp) != 0 {
		t.Errorf("resp: %#v, want %#v", resp, expectedResp)
	}
	// Send the final ACK
	data = []byte{0x00, 0x01}
	resp, err = HandleReadData(data, conn, rAddr, fm)
	if err != nil {
		t.Error(err)
	}
	if resp != nil {
		t.Error("Expected nil response for final ACK")
	}
}

func TestHandleReadDataNoConnInfo(t *testing.T) {
	filename := "foo"
	fileBytes := []byte{0x61, 0x62, 0x63}
	fm := fmgr.NewWithExistingFiles(map[string][]byte{filename: fileBytes})
	ip := "127.0.0.1"
	data := []byte{0x00, 0x00}
	lAddr := &net.UDPAddr{IP: net.ParseIP(ip)}
	conn, err := net.ListenUDP(lAddr.Network(), lAddr)
	if err != nil {
		t.Fatal("Error getting new UDP conn:", err)
	}
	defer conn.Close()
	lAddr = conn.LocalAddr().(*net.UDPAddr)
	rAddr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: lAddr.Port - 1}
	_, err = HandleReadData(data, conn, rAddr, fm)
	if err == nil {
		t.Error("Expected error due to no conn info")
	}
}
