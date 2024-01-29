// Pages 120-121
// Listing 6-1: Types and codes used by the TFTP server.
package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	// TFTP limits datagram packets to 516 bytes or fewer to avoid
	// fragmentation.
	// We define two constant size.
	// The maximum block si ze is the datagram size minus a 4-byte header.
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

// Page 126
// Listing 6-4: Date type and its binary marshaling method.
// Data struct keeps track of the current block number and the data source.
type Data struct {
	Block   uint16
	Payload io.Reader
}

// MarshalBinary will return 516 bytes per call at most by relying on the
// io.CopyN function and the BlockSize constant.
func (d *Data) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(DatagramSize)

	d.Block++ // block numbers increment from 1

	err := binary.Write(b, binary.BigEndian, OpData) // write operation code
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.BigEndian, d.Block) // write block number
	if err != nil {
		return nil, err
	}

	// write up to BlockSize worth of bytes
	_, err = io.CopyN(b, d.Payload, BlockSize)
	if err != nil || err != io.EOF {
		return nil, err
	}

	return b.Bytes(), nil
}

// Page 127
// Listing 6-5: Data type implementation.
func (d *Data) UnmarshalBinary(p []byte) error {
	if l := len(p); l < 4 || l > DatagramSize {
		return errors.New("invalid DATA")
	}

	var opcode OpCode

	// Reads and checks the operation code.
	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &opcode)
	if err != nil || opcode != OpData {
		return errors.New("invalid DATA")
	}

	// Read and checks the block number.
	err = binary.Read(bytes.NewReader(p[2:4]), binary.BigEndian, &d.Block)
	if err != nil {
		return errors.New("invalid DATA")
	}

	// Moves the remaining bytes into a new buffer and assigns it to the Payload field
	d.Payload = bytes.NewBuffer(p[4:])

	return nil
}

// Pages 128 - 129
// Listing 6-6: Acknowledgement type implementation.
type Ack uint16 // Acknowledgement packet represented by a 16-bit, unsigned integer.

func (a Ack) MarshalBinary() ([]byte, error) {
	cap := 2 + 2 // operation code + block number

	b := new(bytes.Buffer)
	b.Grow(cap)

	err := binary.Write(b, binary.BigEndian, OpAck) // write operation code
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.BigEndian, a) // write block number
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (a *Ack) UnmarshalBinary(p []byte) error {
	var code OpCode

	r := bytes.NewReader(p)

	err := binary.Read(r, binary.BigEndian, &code) // read operation code
	if err != nil {
		return err
	}

	if code != OpAck {
		return errors.New("invalid ACK")
	}

	return binary.Read(r, binary.BigEndian, a) // read block number
}

// Page 129 - 130
// Listing 6-7: Error type used for conveying errors between the client and server.
type Err struct {
	Error   ErrCode
	Message string
}

func (e Err) MarshalBinary() ([]byte, error) {
	// operation code + error code + message + 0 byte
	cap := 2 + 2 + len(e.Message) + 1

	b := new(bytes.Buffer)
	b.Grow(cap)

	err := binary.Write(b, binary.BigEndian, OpErr) // write operation code
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.BigEndian, e.Error) // write error code
	if err != nil {
		return nil, err
	}

	_, err = b.WriteString(e.Message) // write message
	if err != nil {
		return nil, err
	}

	err = b.WriteByte(0) // write 0 byte
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// Page 130
// Listing 6-8: Error type's binary unmarshaler implementation.
func (e *Err) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)

	var code OpCode

	err := binary.Read(r, binary.BigEndian, &code) // read operation code
	if err != nil {
		return err
	}

	if code != OpErr {
		return errors.New("invalid ERROR")
	}

	err = binary.Read(r, binary.BigEndian, &e.Error) // read error message
	if err != nil {
		return err
	}

	e.Message, err = r.ReadString(0)
	e.Message = strings.TrimRight(e.Message, "\x00") // remove the 0-byte

	return err
}
