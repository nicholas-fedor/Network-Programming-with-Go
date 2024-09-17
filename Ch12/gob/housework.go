// Pages 278-279
// Listing 12-13: Gob storage implementation
package gob

import (
	"encoding/gob"
	"io"

	"Ch12/housework"
)

func Load(r io.Reader) ([]*housework.Chore, error) {
	var chores []*housework.Chore

	return chores, gob.NewDecoder(r).Decode(&chores)
}

func Flush(w io.Writer, chores []*housework.Chore) error {
	return gob.NewEncoder(w).Encode(chores)
}