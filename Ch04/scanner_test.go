// Page 77
// Listing 4-2: Creating a test to serve up a constant payload.
package main

import (
	"bufio"
	"net"
	"reflect"
	"testing"
)

const payload = "The bigger the interface, the weaker the abstraction."

func TestScanner(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		_, err = conn.Write([]byte(payload))
		if err != nil {
			t.Error(err)
		}
	}()

	// Page 77 - 78
	// Listing 4-3: Using bufio.Scanner to read whitespace-delimited text from
	// the network.
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Creates a bufio.Scanner that reads from the network connection since we
	// know we're reading a string from the server.
	scanner := bufio.NewScanner(conn)

	// By default, the scanner will split data read form the network connection
	// when it encounters a newline character (\n) in the stream of data.
	// Instead, we elect to have the scanner delimit the input at the end of
	// each word by using bufio.ScanWords, which will split the data when it
	// encounters a word border, such as a whitespace or sentence-terminating punctuation.
	scanner.Split(bufio.ScanWords)

	var words []string

	// We keep reading data from the scanner as long as it tells us it's read
	// data from the connection. Every call to Scan can result in multiple calls
	// to the network connection's Read method until the scanner finds its
	// delimiter or reads an error from the connection.
	// It hides the complexity of searching for a delimiter across one or more
	// reads from the network connection and returning hte resulting messages.
	for scanner.Scan() {
		// The call to the scanner's Text method returns the chunk of data as a
		// string - a single word and adjacent punctuation, in this case - that
		// it just read from the network connection.
		// The code continues to iterate around the for loop until the scanner
		// receives an io.EOF or other error from the network connection.
		// If it's the latter, the scanner's Err method will return a non-nil error.
		words = append(words, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		t.Error(err)
	}

	expected := []string{"The", "bigger", "the", "interface,", "the", "weaker", "the", "abstraction."}

	if !reflect.DeepEqual(words, expected) {
		t.Fatal("inaccurate scanned word list")
	}
	// We can view the scanned words by adding the -v flag to the go test command.
	t.Logf("Scanned words: %#v", words)
}
