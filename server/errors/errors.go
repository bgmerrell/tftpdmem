// Package errors provides some server error types
package errors

import (
	"fmt"
)

type UnexpectedRemoteTidErr struct {
	Tid         int
	ExpectedTid int
}

func (e UnexpectedRemoteTidErr) Error() string {
	return fmt.Sprintf("Got remote tid: %d, want %d", e.Tid, e.ExpectedTid)
}

type SrvError struct {
	Code uint16
	Msg  string
}

func (e *SrvError) Error() string {
	return fmt.Sprintf("%s (%d)", e.Msg, e.Code)
}
