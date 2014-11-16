package util

import (
	"bytes"
	"testing"
)

func TestBuildResponse(t *testing.T) {
	data := []interface{}{
		uint16(1024),
		[]byte{'a', 'b', 'c'},
		uint8(56)}
	resp, err := BuildResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0x04, 0x00, 0x61, 0x62, 0x63, 0x38}
	if bytes.Compare(resp, expected) != 0 {
		t.Errorf("Got %#v, want %#v", resp, expected)
	}
}

func TestBuildResponseErr(t *testing.T) {
	data := []interface{}{"fail"}
	_, err := BuildResponse(data)
	if err == nil {
		t.Error("Expected error building response from func")
	}
}
