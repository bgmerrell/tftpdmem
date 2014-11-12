package codes

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
