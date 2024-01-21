// Pages 120-121
// Listing 6-1: Types and codes used by the TFTP server.
package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
)

const (
	// TFTP limits datagram packets to 516 bytes or fewer to avoid
	// fragmentation.
	// We define two constants to enforce the datagram size limit and the
	// maximum data block size.
	// The maximum block size is the datagram size minus a 4-byte header.
	DatagramSize = 516              // the maximum supported datagram size
	BlockSize    = DatagramSize - 4 // the DatagramSize minus a 4-byte header
)

// The first two bytes of a TFTP packet's header is an operation code.
// Each operation code is a 2-byte, unsigned integer.
type OpCode uint16

// Our server supports four operations:
// A read request (RRQ), a data operation, an acknowledgement, and an error.
// Since our server is read-only, we skip the write request (WRQ) definition.
const (
	OpRRQ OpCode = iota + 1
	_            // no WRQ support
	OpData
	OpAck
	OpErr
)

// We define a series of unsigned 16-bit integer error codes per the RFC.
// Although we don't use all error codes in our server since it only allows
// downloads, a client could return these error codes in lieu of an
// acknowledgement packet.
type ErrCode uint16

const (
	ErrUnknown ErrCode = iota
	ErrNotFound
	ErrAccessViolation
	ErrDiskFull
	ErrIllegalOp
	ErrUnknownID
	ErrFileExists
	ErrNoUser
)

// Pages 122-123
// Listing 6-2: Read request and its binary marshaling method.
// This defines the read request and its method that allows the server to
// marshal the request into a slice of bytes in preparation for writing to a
// network connection.

// The struct representing the read request needs to keep track of the filename
// and the mode.
type ReadReq struct {
	Filename string
	Mode     string
}

// Although not used by our server, a client would make use of this method.
func (q ReadReq) MarshalBinary() ([]byte, error) {
	mode := "octet"
	if q.Mode != "" {
		mode = q.Mode
	}

	// operation code + filename + 0 byte + mode + 0 byte
	cap := 2 + 2 + len(q.Filename) + 1 + len(q.Mode) + 1

	b := new(bytes.Buffer)
	b.Grow(cap)

	// Inserts the operation code into the buffer while marshalling the packet
	// into a byte slice.
	err := binary.Write(b, binary.BigEndian, OpRRQ) // write operation code
	if err != nil {
		return nil, err
	}

	_, err = b.WriteString(q.Filename)
	if err != nil {
		return nil, err
	}

	// Inserts null byte into the buffer while marshalling the packet
	// into a byte slice.
	err = b.WriteByte(0) // write 0 byte
	if err != nil {
		return nil, err
	}

	_, err = b.WriteString(mode) // write mode
	if err != nil {
		return nil, err
	}

	// Inserts null byte into the buffer while marshalling the packet
	// into a byte slice.
	err = b.WriteByte(0) // write 0 byte
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// Pages 123-124
// Listing 6-3: Read request type implementation.
// This rounds out the read request's implementation by defining a method that
// allows the server to unmarshal a read request from a byte slice, typically
// read from a network connection with a client.
func (q *ReadReq) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)

	var code OpCode

	// The UnmarshalBinary method reads in the first 2 bytes and confirms the
	// operation code is that of a read request.
	err := binary.Read(r, binary.BigEndian, &code) // read operation code
	if err != nil {
		return nil
	}

	if code != OpRRQ {
		return errors.New("invalid RRQ")
	}

	// It then reads all bytes up to the first null byte.
	q.Filename, err = r.ReadString(0) // read filename
	if err != nil {
		return errors.New("invalid RRQ")
	}

	// Strips the null byte delimiter.
	q.Filename = strings.TrimRight(q.Filename, "\x00") // remove the 0-byte
	if len(q.Filename) == 0 {
		return errors.New("invalid RRQ")
	}

	q.Mode, err = r.ReadString(0) // read mode
	if err != nil {
		return errors.New("invalid RRQ")
	}

	q.Mode = strings.TrimRight(q.Mode, "\x00") // remove the 0-byte
	if len(q.Mode) == 0 {
		return errors.New("invalid RRQ")
	}

	actual := strings.ToLower(q.Mode) // enforce octet mode
	if actual != "octet" {
		return errors.New("only binary transfers supported")
	}

	return nil
}
