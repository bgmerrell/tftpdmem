package filemanager

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bgmerrell/tftpdmem/defs"
	errs "github.com/bgmerrell/tftpdmem/server/errors"
)

func TestFileExistsTrue(t *testing.T) {
	name := "foo"
	tfm := New()
	tfm.filenameToData = map[string][]byte{name: []byte{}}
	exists := tfm.FileExists(name)
	if !exists {
		t.Errorf("Expected filename \"%s\" to exist", name)
	}
}

func TestFileExistsFalse(t *testing.T) {
	name := "foo"
	tfm := New()
	tfm.filenameToData = map[string][]byte{}
	exists := tfm.FileExists(name)
	if exists {
		t.Errorf("Expected filename \"%s\" to not exist", name)
	}
}

func TestAddFile(t *testing.T) {
	name := "foo"
	tfm := New()
	tfm.filenameToData = map[string][]byte{}
	err := tfm.AddFile(name, []byte{'a', 'b', 'c'})
	if err != nil {
		t.Fatal(err)
	}
	exists := tfm.FileExists(name)
	if !exists {
		t.Errorf("Expected filename \"%s\" to exist", name)
	}
	bytes := tfm.filenameToData[name]
	expected := "abc"
	if string(bytes) != expected {
		t.Errorf("File contains %q, want: %q", bytes, expected)
	}
}

func TestAddFileFail(t *testing.T) {
	name := "foo"
	tfm := New()
	tfm.filenameToData = map[string][]byte{}
	err := tfm.AddFile(name, []byte{'a', 'b', 'c'})
	if err != nil {
		t.Fatal(err)
	}
	err = tfm.AddFile(name, []byte{'d', 'e', 'f'})
	if err == nil {
		t.Errorf("Expected error adding \"%s\" for second time", name)
	}
	bytes := tfm.filenameToData[name]
	expected := "abc"
	if string(bytes) != expected {
		t.Errorf("File contains %q, want: %q", bytes, expected)
	}
}

func TestAddConnInfo(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	nextBlockNum := uint16(9)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	expected := &connInfo{filename, remoteTid, nextBlockNum, []byte{}}
	ci := tfm.tidToConnInfo[localTid]
	if ci.filename != expected.filename {
		t.Errorf("filename: %s, want: %s", ci.filename, expected.filename)
	}
	if ci.remoteTid != expected.remoteTid {
		t.Errorf("remote tid: %d, want: %d", ci.remoteTid, expected.remoteTid)
	}
	if ci.nextBlockNum != expected.nextBlockNum {
		t.Errorf("nextBlockNum: %d, want: %d", ci.nextBlockNum, expected.nextBlockNum)
	}
	if bytes.Compare(ci.data, expected.data) != 0 {
		t.Errorf("filename: %#v, want: %#v", ci.data, expected.data)
	}
}

func TestAddConnInfoFail(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	nextBlockNum := uint16(9)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	err = tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err == nil {
		t.Fatal("Expected error adding conn info a second time")
	}
	expected := &connInfo{filename, remoteTid, nextBlockNum, []byte{}}
	ci := tfm.tidToConnInfo[localTid]
	if ci.filename != expected.filename {
		t.Errorf("filename: %s, want: %s", ci.filename, expected.filename)
	}
	if ci.remoteTid != expected.remoteTid {
		t.Errorf("remote tid: %d, want: %d", ci.remoteTid, expected.remoteTid)
	}
	if ci.nextBlockNum != expected.nextBlockNum {
		t.Errorf("nextBlockNum: %d, want: %d", ci.nextBlockNum, expected.nextBlockNum)
	}
	if bytes.Compare(ci.data, expected.data) != 0 {
		t.Errorf("filename: %#v, want: %#v", ci.data, expected.data)
	}
}

func TestDelConnInfo(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	nextBlockNum := uint16(9)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	tfm.DelConnInfo(localTid)
	_, ok := tfm.tidToConnInfo[localTid]
	if ok {
		t.Error("Expected no conn info for local tid:", localTid)
	}
}

func TestWrite(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	inData := []byte("abc")
	nextBlockNum := uint16(9)
	blockNum := uint16(9)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	err = tfm.Write(localTid, remoteTid, blockNum, inData)
	if err != nil {
		t.Error(err)
	}
	outData := tfm.filenameToData[filename]
	if bytes.Compare(inData, outData) != 0 {
		t.Errorf("read: %#v, want: %#v", outData, inData)
	}
}

func TestWriteMultiple(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	blockSizedStr := strings.Repeat("a", defs.BlockSize)
	inData1 := []byte(blockSizedStr)
	inData2 := []byte("test")
	expectedData := []byte(blockSizedStr + "test")
	nextBlockNum := uint16(9)
	blockNum := uint16(9)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	err = tfm.Write(localTid, remoteTid, blockNum, inData1)
	if err != nil {
		t.Error(err)
	}
	blockNum++
	err = tfm.Write(localTid, remoteTid, blockNum, inData2)
	if err != nil {
		t.Error(err)
	}
	outData := tfm.filenameToData[filename]
	if bytes.Compare(outData, expectedData) != 0 {
		t.Errorf("read: %#v, want: %#v", outData, expectedData)
	}
}

func TestWriteNoConnInfo(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	inData := []byte("abc")
	blockNum := uint16(9)
	tfm := New()
	err := tfm.Write(localTid, remoteTid, blockNum, inData)
	if err == nil {
		t.Error("Expected failure writing to file w/o conn info")
	}
}

func TestWriteFileExists(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	inData := []byte("abc")
	nextBlockNum := uint16(9)
	blockNum := uint16(9)
	tfm := New()
	tfm.filenameToData = map[string][]byte{filename: []byte{}}
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	err = tfm.Write(localTid, remoteTid, blockNum, inData)
	if err == nil {
		t.Error("Expected failure due to existing file")
	}
}

func TestWriteBadBlockNum(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	blockSizedStr := strings.Repeat("a", defs.BlockSize)
	inData1 := []byte(blockSizedStr)
	inData2 := []byte("test")
	nextBlockNum := uint16(9)
	blockNum := uint16(9)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	err = tfm.Write(localTid, remoteTid, blockNum, inData1)
	if err != nil {
		t.Error(err)
	}
	err = tfm.Write(localTid, remoteTid, blockNum, inData2)
	if err == nil {
		t.Error("Expected error writing wrong block number")
	}
}

func TestRead(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	inData := []byte("abc")
	blockNum := uint16(0)
	nextBlockNum := uint16(0)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	tfm.filenameToData["foo"] = inData
	outData, err := tfm.Read(localTid, remoteTid, blockNum)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(inData, outData) != 0 {
		t.Errorf("read: %#v, want: %#v", outData, inData)
	}
}

func TestReadNoConnInfo(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	blockNum := uint16(0)
	tfm := New()
	_, err := tfm.Read(localTid, remoteTid, blockNum)
	if err == nil {
		t.Error("Expected failure writing to file w/o conn info")
	}
}

func TestReadWrongRemoteTid(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	remoteTidBad := 5680
	filename := "foo"
	inData := []byte("abc")
	blockNum := uint16(0)
	nextBlockNum := uint16(0)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	tfm.filenameToData["foo"] = inData
	_, err = tfm.Read(localTid, remoteTidBad, blockNum)
	if _, ok := err.(errs.UnexpectedRemoteTidErr); !ok {
		t.Error("Expected UnexpectedRemoteTidErr from mismatched remote tids")
	}
}

func TestReadMultiple(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	blockSizedStr := strings.Repeat("a", defs.BlockSize)
	expectedData1 := []byte(blockSizedStr)
	expectedData2 := []byte("test")
	inData := []byte(blockSizedStr + "test")
	nextBlockNum := uint16(0)
	blockNum := uint16(0)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	tfm.filenameToData[filename] = inData
	outData, err := tfm.Read(localTid, remoteTid, blockNum)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(outData, expectedData1) != 0 {
		t.Errorf("read: %#v, want: %#v", outData, expectedData1)
	}
	blockNum++
	outData, err = tfm.Read(localTid, remoteTid, blockNum)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(outData, expectedData2) != 0 {
		t.Errorf("read: %#v, want: %#v", outData, expectedData2)
	}
	// Simulate the final ACK
	blockNum++
	outData, err = tfm.Read(localTid, remoteTid, blockNum)
	if outData != nil || err != nil {
		t.Errorf("Got outData: %#v, err: %#v.  Want nil for both", outData, err)
	}
}

func TestReadBadBlock(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	filename := "foo"
	blockSizedStr := strings.Repeat("a", defs.BlockSize)
	expectedData1 := []byte(blockSizedStr)
	inData := []byte(blockSizedStr + "test")
	nextBlockNum := uint16(0)
	blockNum := uint16(0)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	tfm.filenameToData[filename] = inData
	outData, err := tfm.Read(localTid, remoteTid, blockNum)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(outData, expectedData1) != 0 {
		t.Errorf("read: %#v, want: %#v", outData, expectedData1)
	}
	_, err = tfm.Read(localTid, remoteTid, blockNum)
	if err == nil {
		t.Error("Expected error reading wrong block number")
	}
}

func TestWriteWrongRemoteTid(t *testing.T) {
	localTid := 1234
	remoteTid := 5678
	remoteTidBad := 5680
	filename := "foo"
	inData := []byte("abc")
	blockNum := uint16(0)
	nextBlockNum := uint16(0)
	tfm := New()
	err := tfm.AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	err = tfm.Write(localTid, remoteTidBad, blockNum, inData)
	if _, ok := err.(errs.UnexpectedRemoteTidErr); !ok {
		t.Error("Expected UnexpectedRemoteTidErr from mismatched remote tids")
	}
}
