package filemanager

import (
	"bytes"
	"testing"
)

func TestFileExistsTrue(t *testing.T) {
	name := "foo"
	filenameToData = map[string][]byte{name: []byte{}}
	exists := FileExists(name)
	if !exists {
		t.Errorf("Expected filename \"%s\" to exist", name)
	}
}

func TestFileExistsFalse(t *testing.T) {
	name := "foo"
	filenameToData = map[string][]byte{}
	exists := FileExists(name)
	if exists {
		t.Errorf("Expected filename \"%s\" to not exist", name)
	}
}

func TestAddFile(t *testing.T) {
	name := "foo"
	filenameToData = map[string][]byte{}
	err := AddFile(name, []byte{'a', 'b', 'c'})
	if err != nil {
		t.Fatal(err)
	}
	exists := FileExists(name)
	if !exists {
		t.Errorf("Expected filename \"%s\" to exist", name)
	}
	bytes := filenameToData[name]
	expected := "abc"
	if string(bytes) != expected {
		t.Errorf("File contains %q, want: %q", bytes, expected)
	}
}

func TestAddFileFail(t *testing.T) {
	name := "foo"
	filenameToData = map[string][]byte{}
	err := AddFile(name, []byte{'a', 'b', 'c'})
	if err != nil {
		t.Fatal(err)
	}
	err = AddFile(name, []byte{'d', 'e', 'f'})
	if err == nil {
		t.Errorf("Expected error adding \"%s\" for second time", name)
	}
	bytes := filenameToData[name]
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
	tidToConnInfo = make(map[int]*connInfo)
	err := AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	expected := &connInfo{filename, remoteTid, nextBlockNum, []byte{}}
	ci := tidToConnInfo[localTid]
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
	tidToConnInfo = make(map[int]*connInfo)
	err := AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err != nil {
		t.Fatal(err)
	}
	err = AddConnInfo(localTid, remoteTid, filename, nextBlockNum)
	if err == nil {
		t.Fatal("Expected error adding conn info a second time")
	}
	expected := &connInfo{filename, remoteTid, nextBlockNum, []byte{}}
	ci := tidToConnInfo[localTid]
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
