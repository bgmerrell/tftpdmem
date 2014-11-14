package filemanager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/bgmerrell/tftpdmem/defs"
)

type UnexpectedRemoteTidErr struct {
	tid         int
	expectedTid int
}

func (e UnexpectedRemoteTidErr) Error() string {
	return fmt.Sprintf("Got remote tid: %d, want %d", e.tid, e.expectedTid)
}

type connInfo struct {
	filename     string
	nextBlockNum uint16
	remoteTid    int
	data         []byte
}

var filenameToData map[string]([]byte) = make(map[string]([]byte))
var fileMu sync.Mutex
var tidToConnInfo map[int]*connInfo = make(map[int]*connInfo)
var connMu sync.Mutex

const firstDataBlock = 1
const fullDataSize = defs.BlockSize - defs.DataHeaderSize

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
		return errors.New(fmt.Sprintf(
			"Filename \"%s\" already exists", filename))
	}
	filenameToData[filename] = data
	return nil
}

// AddConnInfo adds connection info by TID pair
func AddConnInfo(localTid int, remoteTid int, filename string) error {
	connMu.Lock()
	defer connMu.Unlock()
	_, ok := tidToConnInfo[localTid]
	if ok {
		return errors.New(fmt.Sprintf(
			"Local TID %d already exists", localTid))
	}
	tidToConnInfo[localTid] = &connInfo{filename, firstDataBlock, remoteTid, []byte{}}
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
		return UnexpectedRemoteTidErr{remoteTid, info.remoteTid}
	}
	if FileExists(info.filename) {
		DelConnInfo(localTid)
		return errors.New(fmt.Sprintf(
			"Filename \"%s\" already exists", info.filename))
	}
	if blockNum != info.nextBlockNum {
		DelConnInfo(localTid)
		return errors.New(fmt.Sprintf(
			"Got block %d, want %d", blockNum, info.nextBlockNum))
	}
	info.data = append(info.data, buf...)

	// Not done yet...
	if len(buf) == fullDataSize {
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
