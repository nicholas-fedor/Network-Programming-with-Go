package main

import (
	"log"
	"os"
)

func Example_log() {
	// You create a new *log.Logger instance that writes to standard output.
	// The logger prefixes each line with the string "example: ".
	// The flags of the default logger are log.Ldate and log.Ltime, collectively
	// log.LstdFlags, which print the timestamp of each log entry.
	// Since you want to simplify the output for testing purposes when you run
	// the example on the command line, you omit the timestamp and configure the
	// logger to write the source code filename and line of each log entry.
	// The l.Print function on line 12 of the log_test.go file results in the
	// output of those values.
	// This behavior can help with development and debugging, allowing you to
	// zero in on the exact file and line of an interesting log entry.
	l := log.New(os.Stdout, "example: ", log.Lshortfile)
	l.Print("logging to standard output")

	// Output:
	// example: log_test.go:12: logging to standard output
}
