// Page 79
// Listing 4-4: The message struct implements a simple protocol.
package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Creates constants to represent each type we will define.
const (
	BinaryType uint8 = iota + 1
	StringType

	// For security purposes, we define a maximum payload size.
	MaxPayloadSize uint32 = 10 << 20 // 10 MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

// Payload interface describes methods each type must implement.
type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

// Page 80
// Listing 4-5: Creating the Binary type.
// The Binary type is a byte slice.
type Binary []byte

// The Binary type's Bytes method simply returns itself.
func (m Binary) Bytes() []byte { return m }

// The Binary type's String method casts itself as a string before returning.
func (m Binary) String() string { return string(m) }

// The WriteTo method accepts an io.Writer and returns the number of bytes
// written to the writer and an error interface.
func (m Binary) WriteTo(w io.Writer) (int64, error) {
	// The WriteTo method first writes the 1-byte type to the writer.
	err := binary.Write(w, binary.BigEndian, BinaryType) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	// It then writes the 4-byte length of the Binary to the writer.
	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4

	// It then writes the Binary value itself.
	o, err := w.Write(m) // payload

	return n + int64(o), err
}

// Pages 80-81
// Listing 4-6: Completing the Binary type's implementation.
func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	// The ReadFrom method reads 1 byte from the reader into the typ variable.
	err := binary.Read(r, binary.BigEndian, &typ) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	// It next verifies that the type is BinaryType before proceeding.
	if typ != BinaryType {
		return n, errors.New("invalid Binary")
	}

	var size uint32
	// Then it reads the next 4 bytes into the size variable, which sizes the
	// new Binary byte slice.
	err = binary.Read(r, binary.BigEndian, &size) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4
	// We enforce a maximum payload size.
	// This is because the 4-byte integer you use to designate the payload size
	// has a maximum value of 4,294,967,295, indicating a payload of over 4 GB.
	// With such a large payload size, it would be easy for a malicious actor to
	// perform a denial-of-service attack that exhausts all the available RAM on
	// your computer.
	// Keeping the maximum payload size reasonable makes memory exhaustion
	// attacks harder to execute.
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	*m = make([]byte, size)
	// Finally, it populates the Binary byte slice.
	o, err := r.Read(*m) // payload

	return n + int64(o), err
}

// Pages 81-82
// Listing 4-7: Creating the String type.
type String string

func (m String) Bytes() []byte  { return []byte(m) }
func (m String) String() string { return string(m) }

func (m String) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, StringType) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4

	o, err := w.Write([]byte(m)) // payload

	return n + int64(o), err
}

// Page 82
// Listing 4-8: Completing the String type's implementation.
func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	if typ != StringType {
		return n, errors.New("invalid String")
	}

	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // 4-byte size
	if err != nil {
		return n, err
	}
	n += 4

	buf := make([]byte, size)
	o, err := r.Read(buf) // payload
	if err != nil {
		return n, err
	}
	*m = String(buf)

	return n + int64(o), nil
}

// Page 83
// Listing 4-9: Decoding bytes from a reader into a Binary or String type.
// The decode function accepts an io.Reader and returns a Payload interface and
// an error interface.
// If decode cannot decode the bytes read form the reader into a Binary or
// String type, it will return an error along with a nil Payload.
func decode(r io.Reader) (Payload, error) {
	var typ uint8
	// We first read a byte from the reader to determine the type.
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}

	// Creates a payload variable to hold the decoded type.
	var payload Payload

	// If the type you read from the reader is an expected type constant, you
	// assign the corresponding type to the payload variable.
	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType:
		payload = new(String)
	default:
		return nil, errors.New("unknown type")
	}

	_, err = payload.ReadFrom(
		// We use io.MultiReader to concatenate the byte we already read with
		// the reader.
		// This is not an optimal use-case.
		// The proper fix is to remove each type's need to read the first byte
		// in its ReadFrom method.
		// Then, the ReadFrom method would read only the 4-byte size and the
		// payload, eliminating the need to inject the type byte back into the
		// reader before passing it on to ReadFrom.
		io.MultiReader(bytes.NewReader([]byte{typ}), r))
	if err != nil {
		return nil, err
	}

	return payload, nil
}