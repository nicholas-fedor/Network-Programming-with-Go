// Page 283
// Listing 12-18: Protocol buffers storage implementation
package protobuf

import (
	"io"

	"google.golang.org/protobuf/proto"

	// Instead of relying on the housework package from Listing 12-1, as you did
	// when working with JSON and Gob, you import version 1 of the
	// protoc-generated package, which you also named housework.
	"Ch12/housework/v1"
)

func Load(r io.Reader) ([]*housework.Chore, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var chores housework.Chores

	return chores.Chores, proto.Unmarshal(b, &chores)
}

func Flush(w io.Writer, chores []*housework.Chore) error {
	// The generated Chores type is a struct with a Chores field, which itself
	// is a slice of Chore pointers.
	b, err := proto.Marshal(&housework.Chores{Chores: chores})
	if err != nil {
		return err
	}

	// Also, Go's protocol buffers package doesn't implement encoders and
	// decoders. Therefore, you must marshall objects to bytes, write them to
	// the io.Writer, and unmarshal bytes from the io.Reader yourself.
	_, err = w.Write(b)

	return err
}
