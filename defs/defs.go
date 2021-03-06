// The defs package contains misc definitions derived from the TFTP RFC
package defs

// misc
const (
	BlockSize      = 512
	OpCodeSize     = 2
	BlockNumSize   = 2
	DataHeaderSize = OpCodeSize + BlockNumSize
	DatagramSize   = BlockSize + DataHeaderSize
	MinOpCode      = 1
	MaxOpCode      = 5
	FirstDataBlock = 1
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
