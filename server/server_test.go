package server

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/bgmerrell/tftpdmem/defs"
	fmgr "github.com/bgmerrell/tftpdmem/filemanager"
)

// A testServer is a Server that includes a remote server UDPConn so that it
// can send requests to itself.
type testServer struct {
	*Server
	rConn *net.UDPConn
}

func (ts *testServer) Close() {
	ts.Server.conn.Close()
	ts.rConn.Close()
}

func getTestServer(opToHandle OpToHandleMap) (*testServer, error) {
	lAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1")}
	lConn, err := net.ListenUDP(lAddr.Network(), lAddr)
	if err != nil {
		log.Println("ListenUDP failure", err)
		return nil, err
	}
	lAddr = lConn.LocalAddr().(*net.UDPAddr)

	rAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1")}
	rConn, err := net.ListenUDP(rAddr.Network(), rAddr)
	if err != nil {
		log.Println("ListenUDP failure", err)
		return nil, err
	}
	rAddr = rConn.LocalAddr().(*net.UDPAddr)

	return &testServer{
		Server: New(lAddr.Port, lConn, opToHandle, false, fmgr.New()),
		rConn:  rConn}, err
}

func TestServe(t *testing.T) {
	opToHandle := OpToHandleMap{
		defs.OpWrq: func([]byte, *net.UDPConn, *net.UDPAddr, *fmgr.FileManager) ([]byte, error) { return []byte("wrq"), nil }}
	readTimeout = 100 * time.Millisecond
	s, err := getTestServer(opToHandle)
	if err != nil {
		t.Fatal("Failed to get test server:", err)
	}
	go s.Serve()
	rrqData := []byte{0x00, 0x01}
	n, err := s.rConn.WriteToUDP(rrqData, s.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(rrqData) {
		t.Errorf("Read %d bytes, want: %d", n, len(rrqData))
	}
	time.Sleep(200 * time.Millisecond)
	s.StopCh <- struct{}{}
}
