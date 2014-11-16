package handlers

import (
	"bytes"
	"fmt"
	"net"
	"testing"

	"github.com/bgmerrell/tftpdmem/defs"
	fmgr "github.com/bgmerrell/tftpdmem/filemanager"
)

func TestHandleWriteRequest(t *testing.T) {
	fm := fmgr.New()
	ip := "127.0.0.1"
	// Expect an ack with block number set to 0
	expectedData := []byte{0x00, 0x04, 0x00, 0x00}
	laddr := net.UDPAddr{IP: net.ParseIP(ip)}
	conn, err := net.ListenUDP(laddr.Network(), &laddr)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if err != nil {
		t.Fatal("Failed to get UDP conn:", err)
	}
	resp, err := HandleWriteRequest(
		// foo\0octet\0
		[]byte{0x66, 0x6f, 0x6f, 0x00, 0x6f, 0x63, 0x74, 0x65, 0x74, 0x00},
		conn,
		localAddr,
		fm)
	if err != nil {
		t.Fatal(err)
	}
	if resp != nil {
		// The request handles don't respond over the same connection
		t.Errorf("Resp: %#v, want nil", resp)
	}
	buf := make([]byte, defs.DatagramSize)
	fmt.Println("Reading from UDP...")
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(expectedData) {
		t.Errorf("Read %d bytes, want: %d", n, len(expectedData))
	}
	if addr.IP.String() != ip {
		t.Errorf("IP: %s, want: %s", addr.IP, ip)
	}
	// request handle should respond from a different port
	if addr.Port == localAddr.Port {
		t.Errorf("Expected port (%d) to change", addr.Port)
	}
	if bytes.Compare(buf[:n], expectedData) != 0 {
		t.Errorf("Data: %#v, want: %#v", buf, expectedData)
	}
}

func TestHandleReadRequest(t *testing.T) {
	filename := "foo"
	// abc
	fileBytes := []byte{0x61, 0x62, 0x63}
	fm := fmgr.NewWithExistingFiles(map[string][]byte{filename: fileBytes})
	// Data packet with block num 1 and the file bytes
	expectedData := []byte{0x00, 0x03, 0x00, 0x01, 0x61, 0x62, 0x63}
	ip := "127.0.0.1"
	laddr := net.UDPAddr{IP: net.ParseIP(ip)}
	conn, err := net.ListenUDP(laddr.Network(), &laddr)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if err != nil {
		t.Fatal("Failed to get UDP conn:", err)
	}
	resp, err := HandleReadRequest(
		// foo\0octet\0
		[]byte{0x66, 0x6f, 0x6f, 0x00, 0x6f, 0x63, 0x74, 0x65, 0x74, 0x00},
		conn,
		localAddr,
		fm)
	if err != nil {
		t.Fatal(err)
	}
	if resp != nil {
		// The request handles don't respond over the same connection
		t.Errorf("Resp: %#v, want nil", resp)
	}
	buf := make([]byte, defs.DatagramSize)
	fmt.Println("Reading from UDP...")
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(expectedData) {
		t.Errorf("Read %d bytes, want: %d", n, len(expectedData))
	}
	if addr.IP.String() != ip {
		t.Errorf("IP: %s, want: %s", addr.IP, ip)
	}
	// request handle should respond from a different port
	if addr.Port == localAddr.Port {
		t.Errorf("Expected port (%d) to change", addr.Port)
	}
	if bytes.Compare(buf[:n], expectedData) != 0 {
		t.Errorf("Data: %#v, want: %#v", buf, expectedData)
	}
}
