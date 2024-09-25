package Ch13

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

// Page 297
// Listing 13-1: Writing a long entry to standard output.
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
	// example: log_test.go:25: logging to standard output
}

// Page 299
// Listing 13-4: Logging simultaneously to a log file and standard output.
func Example_logMultiWriter()  {
	logFile := new(bytes.Buffer)
	// You create a new sustained multiwriter, writing to standard output, and a
	// bytes.Buffer meant to act as a log file in this example.
	w := SustainedMultiWriter(os.Stdout, logFile)
	// Next, you create a new logger using your sustained multiwriter, the
	// prefix example:, and two flags to modify the logger's behavior.
	// The addition of the log.Lmsgprefix flag tells the logger to locate the
	// prefix just before the log message.
	// You can see the effect this has on the log entries in the example output.
	l := log.New(w, "example: ", log.Lshortfile|log.Lmsgprefix)
	
	// When you run this example, you see the logger writes the log entry to the
	// sustained multiwriter, which in turns writes the log entry to both
	// standard output and the log file.
	fmt.Println("standard output:")
	l.Print("Canada is south of Detroit")

	fmt.Print("\nlog file contents:\n", logFile.String())

	// Output:
	// standard output:
	// log_test.go:49: example: Canada is south of Detroit
	//
	// log file contents:
	// log_test.go:49: example: Canada is south of Detroit
}

// Page 300
// Listing 13-5: Writing debug entries to standard output and errors to both the
// log file and standard output.
func Example_logLevels()  {
	// First, you create a debug logger that writes to standard output and uses
	// the DEBUG: prefix.
	lDebug := log.New(os.Stdout, "DEBUG: ", log.Lshortfile)
	// Next, you create a *bytes.Buffer to masquerade as a log file for this
	// example and to instantiate a sustained multiwriter.
	logFile := new(bytes.Buffer)
	// The sustained multiwriter writes to both the log file and the debug
	// loggers's io.Writer.
	w := SustainedMultiWriter(logFile, lDebug.Writer())
	// Then, you create an error logger that writes to the sustained multiwriter
	// by using the prefix ERROR: to differentiate its log entries from the
	// debug logger.
	lError := log.New(w, "ERROR: ", log.Lshortfile)

	// Finally, you use each logger and verify that the output what you expect.
	// Standard output should display log entries from both loggers, whereas the
	// log file should contain only log entries.
	fmt.Println("standard output:")
	lError.Print("cannot communicate with the database")
	lDebug.Print("you cannot hum while holding your nose")

	fmt.Print("\nlog file contents:\n", logFile.String())

	// Output:
	// standard output:
	// ERROR: log_test.go:83: cannot communicate with the database
	// DEBUG: log_test.go:84: you cannot hum while holding your nose
	//
	// log file contents:
	// ERROR: log_test.go:83: cannot communicate with the database
}