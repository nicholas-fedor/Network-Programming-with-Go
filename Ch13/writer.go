// Pages 298
package Ch13

import (
	"io"

	"go.uber.org/multierr"
)

// Recognizing that the logger accepts an io.Writer, you may realize this allows
// you to use multiple writers, such as a log file and standard output or an
// in-memory ring buffer and a centralized logging server over a network.
// Unfortunately, the io.MultiWriter is not ideal for use in logging.
// An io.MultiWriter writes to each of its writers in sequence, aborting if it
// receives an error from any Write call.
// This means that if you configure your io.MultiWriter to write to a log file
// and standard output in that order, standard output will never receive the log
// entry if an error occurred when writing to the log file.

type sustainedMultiWriter struct {
	writers []io.Writer
}

// Page 298
// Listing 13-2: A multiwriter that sustains writing even after receiving an
// error As with io.MultiWriter, you'll use a struct that contains a slice of
// io.Writer instances for your sustained multiwriter. Your multiwriter
// implements the io.Writer interface, so you can pass it into your logger.
func (s *sustainedMultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range s.writers {
		// It calls each writer's Write method
		i, wErr := w.Write(p)
		n += i
		// accumulating any errors with the help of Uber's multierr package,
		// before ultimately returning the total written bytes and cumulative
		// error.
		err = multierr.Append(err, wErr)
	}

	return n, err
}

// Pages 298-299
// Listing 13-3: Creating a sustained multiwriter
func SustainedMultiWriter(writers ...io.Writer) io.Writer {
	// First, you instantiate a new *sustainedMultiWriter, initialize its
	// writers slice, and cap it to the expected length of writers.
	mw := &sustainedMultiWriter{writers: make([]io.Writer, 0, len(writers))}

	for _, w := range writers {
		// If a given writer is itself a *sustainedMultiWriter...
		if m, ok := w.(*sustainedMultiWriter); ok {
			// ...you append its writers.
			mw.writers = append(mw.writers, m.writers...)
			continue
		}

		// You then loop through the given writers and append them to the slice.
		mw.writers = append(mw.writers, w)
	}

	// Finally, you return the pointer to the initialized sustainedMultiWriter.
	return mw
}