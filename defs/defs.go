// The defs package contains misc definitions derived from the TFTP RFC
package defs

// misc
const (
	BlockSize = 512
	MinOpCode = 1
	MaxOpCode = 5
)

// op codes
const (
	OpRrq = iota + 1
	OpWrq
	OpData
	OpAck
	OpErr
)

// err codes
const (
	ErrGeneric = iota
	ErrFileNotFound
	ErrAccessViolation
	ErrFull
	ErrIllegalOp
	ErrUnknownTid
	ErrFileExists
	ErrNoSuchUser
)
