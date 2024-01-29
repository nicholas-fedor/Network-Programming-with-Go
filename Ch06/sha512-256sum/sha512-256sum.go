// Page 137
// Listing 6-12: Generating SHA512/256 checksums for given command line
// arguments.
// This will accept one or more file paths as command-line arguments and
// generate SHA512/256 checksums from their contents.
package main

import (
	"crypto/sha512"
	"flag"
	"fmt"
	"os"
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s file...\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	for _, file := range flag.Args() {
		fmt.Printf("%s   %s\n", checksum(file), file)
	}
}

func checksum(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%x", sha512.Sum512_256(b))
}
