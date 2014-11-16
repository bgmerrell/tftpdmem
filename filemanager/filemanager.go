package filemanager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/bgmerrell/tftpdmem/defs"
	errs "github.com/bgmerrell/tftpdmem/server/errors"
)

type FileManager struct {
	filenameToData map[string]([]byte)
	fileMu         sync.Mutex
	tidToConnInfo  map[int]*connInfo
	connMu         sync.Mutex
}

type connInfo struct {
	filename     string
	remoteTid    int
	nextBlockNum uint16
	data         []byte
}

// New returns a new FileManager.
func New() *FileManager {
	return &FileManager{
		filenameToData: make(map[string]([]byte)),
		tidToConnInfo:  make(map[int]*connInfo)}
}

// NewWithExistingFiles returns a FileManager with prepopulated files.  Handy
// for testing.
func NewWithExistingFiles(filenameToData map[string]([]byte)) *FileManager {
	return &FileManager{
		filenameToData: filenameToData,
		tidToConnInfo:  make(map[int]*connInfo)}
}

// FileExists returns whether or not a file exists
func (fm *FileManager) FileExists(filename string) bool {
	fm.fileMu.Lock()
	defer fm.fileMu.Unlock()
	_, ok := fm.filenameToData[filename]
	return ok
}

// AddFile adds a new file with data
func (fm *FileManager) AddFile(filename string, data []byte) error {
	fm.fileMu.Lock()
	defer fm.fileMu.Unlock()
	_, ok := fm.filenameToData[filename]
	if ok {
		return &errs.SrvError{defs.ErrFileExists,
			fmt.Sprintf("Filename \"%s\" already exists", filename)}
	}
	fm.filenameToData[filename] = data
	return nil
}

// AddConnInfo adds connection info by TID pair
func (fm *FileManager) AddConnInfo(localTid int, remoteTid int, filename string, nextBlockNum uint16) error {
	fm.connMu.Lock()
	defer fm.connMu.Unlock()
	_, ok := fm.tidToConnInfo[localTid]
	if ok {
		return errors.New(fmt.Sprintf(
			"Local TID %d already exists", localTid))
	}
	fm.tidToConnInfo[localTid] = &connInfo{
		filename, remoteTid, nextBlockNum, []byte{}}
	return nil
}

// DelConnInfo deletes connection info by TID pair
func (fm *FileManager) DelConnInfo(localTid int) {
	fm.connMu.Lock()
	defer fm.connMu.Unlock()
	delete(fm.tidToConnInfo, localTid)
}

// Write takes a tid and a blockNum and attempts to write data to a "file"
// buffer.
func (fm *FileManager) Write(localTid int, remoteTid int, blockNum uint16, buf []byte) error {
	fm.connMu.Lock()
	info, ok := fm.tidToConnInfo[localTid]
	fm.connMu.Unlock()
	if !ok {
		return errors.New(fmt.Sprintf(
			"No connection info for local TID (%d)", localTid))
	}
	if remoteTid != info.remoteTid {
		return errs.UnexpectedRemoteTidErr{remoteTid, info.remoteTid}
	}
	if fm.FileExists(info.filename) {
		fm.DelConnInfo(localTid)
		return &errs.SrvError{defs.ErrFileExists,
			fmt.Sprintf("Filename \"%s\" already exists", info.filename)}
	}
	if blockNum != info.nextBlockNum {
		fm.DelConnInfo(localTid)
		return errors.New(fmt.Sprintf(
			"Got block %d, want %d", blockNum, info.nextBlockNum))
	}
	info.data = append(info.data, buf...)

	// Not done yet...
	if len(buf) == defs.BlockSize {
		info.nextBlockNum++
		return nil
	}

	// Done
	err := fm.AddFile(info.filename, info.data)
	if err != nil {
		fm.DelConnInfo(localTid)
		return err
	}
	return nil
}

// Read takes a tid and a blockNum and attempts to read data from a "file"
// buffer.
func (fm *FileManager) Read(localTid int, remoteTid int, blockNum uint16) ([]byte, error) {
	fm.connMu.Lock()
	info, ok := fm.tidToConnInfo[localTid]
	fm.connMu.Unlock()
	if !ok {
		return nil, errors.New(fmt.Sprintf(
			"No connection info for local TID (%d)", localTid))
	}
	if remoteTid != info.remoteTid {
		return nil, errs.UnexpectedRemoteTidErr{remoteTid, info.remoteTid}
	}
	if blockNum != info.nextBlockNum {
		fm.DelConnInfo(localTid)
		return nil, errors.New(fmt.Sprintf(
			"Got block %d, want %d", blockNum, info.nextBlockNum))
	}
	data := fm.filenameToData[info.filename]
	startIdx := int(blockNum * defs.BlockSize)
	endIdx := int(startIdx + defs.BlockSize)
	// A final ACK will put the startIdx out of bounds, and we don't need
	// to respond to it.
	if startIdx > len(data) {
		fm.DelConnInfo(localTid)
		return nil, nil
	} else if endIdx > len(data) {
		endIdx = len(data)
	}
	info.nextBlockNum++

	return data[startIdx:endIdx], nil
}
