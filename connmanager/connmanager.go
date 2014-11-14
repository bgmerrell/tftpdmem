package connmanager

import (
	"fmt"
	"sync"

	"github.com/bgmerrell/tftpdmem/filemanager"
)

type info struct {
	filename string
	blockNum int
}

type ConnManager struct {
	tidToInfo map[int]*info
	fm        *filemanager.FileManager
	mu        sync.Mutex
}

func New() *ConnManager {
	return &ConnManager{
		tidToInfo: make(map[int]*info),
		fm:        filemanager.New()}
}

// Add adds a connection by TID
func (c *conn) Add(tid int, filename string) error {
	c.mu.Lock()
	defer c.mu.Lock()
	_, ok := tidToInfo[tid]
	if ok {
		return errors.New(fmt.Sprintf("TID already exists: %d", tid))
	}
	tidToInfo[tid] = &info{filename, 0}
}

// Reset resets a connection by TID
func (c *conn) Reset(tid int, filename string) error {
	c.mu.Lock()
	defer c.mu.Lock()
	_, ok := tidToInfo[tid]
	if ok {
		return errors.New(fmt.Sprintf("TID already exists: %d", tid))
	}
	tidToInfo[tid] = &info{filename, 0}
}
