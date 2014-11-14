package filemanager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/bgmerrell/tftpdmem/defs"
	errs "github.com/bgmerrell/tftpdmem/server/errors"
)

type connInfo struct {
	filename     string
	remoteTid    int
	nextBlockNum uint16
	data         []byte
}

var filenameToData map[string]([]byte) = make(map[string]([]byte))
var fileMu sync.Mutex
var tidToConnInfo map[int]*connInfo = make(map[int]*connInfo)
var connMu sync.Mutex

// FileExists returns whether or not a file exists
func FileExists(filename string) bool {
	fileMu.Lock()
	defer fileMu.Unlock()
	_, ok := filenameToData[filename]
	return ok
}

// AddFile adds a new file with data
func AddFile(filename string, data []byte) error {
	fileMu.Lock()
	defer fileMu.Unlock()
	_, ok := filenameToData[filename]
	if ok {
		return &errs.SrvError{defs.ErrFileExists,
			fmt.Sprintf("Filename \"%s\" already exists", filename)}
	}
	filenameToData[filename] = data
	return nil
}

// AddConnInfo adds connection info by TID pair
func AddConnInfo(localTid int, remoteTid int, filename string, nextBlockNum uint16) error {
	connMu.Lock()
	defer connMu.Unlock()
	_, ok := tidToConnInfo[localTid]
	if ok {
		return errors.New(fmt.Sprintf(
			"Local TID %d already exists", localTid))
	}
	tidToConnInfo[localTid] = &connInfo{
		filename, remoteTid, nextBlockNum, []byte{}}
	return nil
}

// DelConnInfo deletes connection info by TID pair
func DelConnInfo(localTid int) {
	connMu.Lock()
	defer connMu.Unlock()
	delete(tidToConnInfo, localTid)
}

// Write takes a tid and a blockNum and attempts to write data to a "file"
// buffer.
func Write(localTid int, remoteTid int, blockNum uint16, buf []byte) error {
	connMu.Lock()
	info, ok := tidToConnInfo[localTid]
	connMu.Unlock()
	if !ok {
		return errors.New(fmt.Sprintf(
			"No connection info for local TID (%d)", localTid))
	}
	if remoteTid != info.remoteTid {
		return errs.UnexpectedRemoteTidErr{remoteTid, info.remoteTid}
	}
	if FileExists(info.filename) {
		DelConnInfo(localTid)
		return &errs.SrvError{defs.ErrFileExists,
			fmt.Sprintf("Filename \"%s\" already exists", info.filename)}
	}
	if blockNum != info.nextBlockNum {
		DelConnInfo(localTid)
		return errors.New(fmt.Sprintf(
			"Got block %d, want %d", blockNum, info.nextBlockNum))
	}
	info.data = append(info.data, buf...)
	fmt.Println(len(info.data))

	// Not done yet...
	if len(buf) == defs.BlockSize {
		info.nextBlockNum++
		return nil
	}

	// Done
	err := AddFile(info.filename, info.data)
	if err != nil {
		DelConnInfo(localTid)
		return err
	}
	return nil
}

// Read takes a tid and a blockNum and attempts to read data from a "file"
// buffer.
func Read(localTid int, remoteTid int, blockNum uint16) ([]byte, error) {
	connMu.Lock()
	info, ok := tidToConnInfo[localTid]
	connMu.Unlock()
	if !ok {
		return nil, errors.New(fmt.Sprintf(
			"No connection info for local TID (%d)", localTid))
	}
	if remoteTid != info.remoteTid {
		return nil, errs.UnexpectedRemoteTidErr{remoteTid, info.remoteTid}
	}
	if blockNum != info.nextBlockNum {
		DelConnInfo(localTid)
		return nil, errors.New(fmt.Sprintf(
			"Got block %d, want %d", blockNum, info.nextBlockNum))
	}
	data := filenameToData[info.filename]
	startIdx := int(blockNum * defs.BlockSize)
	endIdx := int(startIdx + defs.BlockSize)
	// A final ACK will put the startIdx out of bounds, and we don't need
	// to respond to it.
	if startIdx > len(data) {
		DelConnInfo(localTid)
		return nil, nil
	} else if endIdx > len(data) {
		endIdx = len(data)
	}
	info.nextBlockNum++

	return data[startIdx:endIdx], nil
}
