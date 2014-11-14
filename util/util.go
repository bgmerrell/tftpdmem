// The util package contains any general purpose helper functions.
package util

import (
	"bytes"
	"encoding/binary"
)

func BuildResponse(data []interface{}) ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}
	for _, v := range data {
		err := binary.Write(buf, binary.BigEndian, v)
		if err != nil {
			break
		}
	}
	return buf.Bytes(), err
}
