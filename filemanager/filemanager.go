package filemanager

import (
	"errors"
	"fmt"
	"sync"
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

var tidToConnInfo map[int]*connInfo = make(map[int]*connInfo)
var mu sync.Mutex

const firstDataBlock = 1

// AddConnInfo adds connection info by TID pair
func AddConnInfo(localTid int, remoteTid int, filename string) (err error) {
	mu.Lock()
	defer mu.Unlock()
	_, ok := tidToConnInfo[localTid]
	if ok {
		return errors.New(fmt.Sprintf(
			"Local TID %d already exists", localTid))
	}
	tidToConnInfo[localTid] = &connInfo{filename, firstDataBlock, remoteTid, []byte{}}
	return err
}

// DelConnInfo deletes connection info by TID pair
func DelConnInfo(localTid int) {
	mu.Lock()
	defer mu.Unlock()
	delete(tidToConnInfo, localTid)
}

// Write takes a tid and a blockNum and attempts to write data to a "file"
// buffer.
func Write(localTid int, remoteTid int, blockNum uint16, buf []byte) error {
	mu.Lock()
	defer mu.Unlock()
	info, ok := tidToConnInfo[localTid]
	if !ok {
		return errors.New(fmt.Sprintf(
			"No connection info for local TID (%d)", localTid))
	}
	if remoteTid != info.remoteTid {
		return UnexpectedRemoteTidErr{remoteTid, info.remoteTid}
	}
	if blockNum != info.nextBlockNum {
		DelConnInfo(localTid)
		return errors.New(fmt.Sprintf(
			"Got block %d, want %d", blockNum, info.nextBlockNum))
	}
	info.nextBlockNum++
	return nil
}
