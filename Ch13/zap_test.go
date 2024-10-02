package Ch13

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/fsnotify.v1"
)

// Pages 302-303
// Listing 13-6: The encoder configuration for your Zap logger.
// The encoder configuration is independent of the encoder itself in that you
// can use the same encoder configuration no matter whether you're passing it to
// a JSON encoder or a console encoder.
// The encoder will use your configuration to dictate its output format.
var encoderCfg = zapcore.EncoderConfig{
	// Here, your encoder configuration dictates tha the encoder will use the
	// key msg for the log message...
	MessageKey: "msg",
	// ...and the key name for the logger's name in the log entry.
	NameKey: "name",

	LevelKey: "level",
	// Likewise, the encoder configuration tells the encoder to use the key
	// level for the logging level and encode the level name using all lowercase characters.
	EncodeLevel: zapcore.LowercaseLevelEncoder,

	// If the logger is configured to add caller details, you want the encoder
	// to associate these details with the caller key and encode the details in
	// an abbreviated format.
	CallerKey:    "caller",
	EncodeCaller: zapcore.ShortCallerEncoder,

	// Since you need to keep the output of the following examples consistent,
	// you'll omit the time key so it won't show up in the output.
	// In practice, you'd want to uncomment these two fields.
	// TimeKey: "time",
	// EncodeTime: zapcore.ISO8601TimeEncoder,
}

// Pages 303-304
// Listing 13-7:
func Example_zapJSON() {
	zl := zap.New(
		// The zap.New function accepts a zap.Core and zero or more zap.Options.
		zapcore.NewCore(
			// The zap.Core consists of a JSON encoder using your encoder configuration,
			zapcore.NewJSONEncoder(encoderCfg),
			// a zapcore.WriteSyncer,
			zapcore.Lock(os.Stdout),
			// and the logging threshold.
			zapcore.DebugLevel,
			// If the zapcore.WriteSyncer isn't safe for concurrent use, you can
			// use zapcore.Lock to make it concurrency safe, as in this example.
			//
			// The Zap logger includes seven log levels in increasing severity:
			// DebugLevel, InfoLevel, WarnLevel, ErrorLevel, DPanicLevel,
			// PanicLevel, and FatalLevel.
			// The InfoLevel is the default.
			// DPanicLevel and PanicLevel entries will cause Zap to log the
			// entry and then panic.
			// An entry logged at the FatalLevel will cause Zap to call
			// os.Exit(1) after writing the log entry.
			// Since your logger is using DebugLevel, it will log all entries.
			//
			// DPanicLevel and PanicLevel are recommended to be restricted to
			// development and FatalLevel to production.
		),
		// In this example, you're passing the zap.AddCaller option, which
		// instructs the logger to include the caller information in each log
		// entry...
		zap.AddCaller(),
		zap.Fields(
			// ...and a field named version that inserts the runtime version in
			// each log entry.
			zap.String("version", runtime.Version()),
		),
	)
	// Before you start using the logger, you want to make sure you defer a call
	// to its Sync method to ensure all buffered data is written to the output.
	defer func() { _ = zl.Sync() }()

	// You can also assign the logger a name by calling its Named method and
	// using hte returned logger.
	// By default, a logger has no name.
	// A named logger will include a name key in the log entry, provided you use
	// one in the encoder configuration.
	example := zl.Named("example")
	example.Debug("test debug message")
	example.Info("test info message")

	// The log entries now include metadata around the log message, so much so
	// that the log line exceeds normal column 80 width.
	// The Go version is dependent on the version of Go being used to test this example.

	// Output:
	// {"level":"debug","name":"example","caller":"Ch13/zap_test.go:90","msg":"test debug message","version":"go1.23.1"}
	// {"level":"info","name":"example","caller":"Ch13/zap_test.go:91","msg":"test info message","version":"go1.23.1"}
}

// Page 305
// Listing 13-8: Writing structured logs using console encoding.
func Example_zapConsole() {
	zl := zap.New(
		zapcore.NewCore(
			// The console encoder uses tabs to separate fields.
			// It takes instruction from your encoder configuration to determine
			// which fields to include and how to format each.
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zapcore.InfoLevel,
			// Notice that you don't pass the zap.AddCaller and zap.Fields
			// options to the logger in this example.
			// As a result, the log entries won't have caller and version
			// fields.
			// Log entries will include the caller field only if the logger has
			// the zap.AddCaller option and the encoder configuration defines
			// its CallerKey, as in Listing 13-6.
		),
	)
	defer func() { _ = zl.Sync() }()

	// You name the logger and write three log entries, each with a different
	// log level.
	console := zl.Named("[console]")
	console.Info("this is logged by the logger")
	// Since the logger's threshold is the info level, the debug log entry does
	// not appear in the output because debug is below the info threshold.
	console.Debug("this is below the logger's threshold and won't log")
	console.Error("this is also logged by the logger")

	// The output lacks key names but includes the field values delimited by a
	// tab character.
	// Output:
	// info	[console]	this is logged by the logger
	// error	[console]	this is also logged by the logger
}

// Page 306
// Listing 13-9: Using *bytes.Buffer as the log output and logging JSON to it.
func Example_zapInfoFileDebugConsole() {
	// You're using *bytes.Buffer to act as a mock log file.
	// The only problem with this is that *bytes.Buffer does not have a Sync
	// method and does not implement the zapcore.WriteSyncer interface.
	logFile := new(bytes.Buffer)
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			// Thankfully, Zap includes a helper function named zapcore.AddSync
			// that intelligently adds a no-op Sync method to an io.Writer.
			// Aside from the use of this function, the rest of the logger
			// implementation should be familiar to you.
			zapcore.Lock(zapcore.AddSync(logFile)),
			// It's logging JSON to the log file and excluding any log entries
			// below the info level.
			zapcore.InfoLevel,
		),
	)
	defer func() { _ = zl.Sync() }()

	// As a result, the first log entry should not appear in the log
	// file at all.
	zl.Debug("this is below the logger's threshold and won't log")
	zl.Error("this is logged by the logger")

	// Pages 306-307
	// Listing 13-10: Extending the logger to log to multiple outputs.
	// Zap's WithOptions method clones the existing logger and configures the
	// clone with the given options.
	zl = zl.WithOptions(
		// You can use the zap.WrapCore function to modify the underlying
		// zap.Core of the cloned logger.
		zap.WrapCore(
			func(c zapcore.Core) zapcore.Core {
				ucEncoderCfg := encoderCfg
				// To mix things up, you make a copy of the encoder
				// configuration and tweak it to instruct the zapcore.NewTee
				// function, which is like the io.MultiWriter function, to
				// return a zap.Core that writes to multiple cores.
				ucEncoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
				return zapcore.NewTee(
					c,
					// In this example, you're passing in the existing core and a
					// new core that writes debug-level log entries to the standard output.
					zapcore.NewCore(
						zapcore.NewConsoleEncoder(ucEncoderCfg),
						zapcore.Lock(os.Stdout),
						zapcore.DebugLevel,
					),
				)
			},
		),
	)

	fmt.Println("standard output:")
	// When you use the cloned logger, both the log file and standard output
	// receive any log entry at the info level or above, whereas only standard
	// output receives debugging log entries.
	zl.Debug("this is only logged as console encoding")
	zl.Info("this is logged as console encoding and JSON")

	fmt.Print("\nlog file contents:\n", logFile.String())

	// standard output:
	// DEBUG	this is only logged as console encoding
	// INFO	this is logged as console encoding and JSON

	// log file contents:
	// {"level":"error","msg":"this is logged by the logger"}
	// {"level":"info","msg":"this is logged as console encoding and JSON"}
}

// Pages 307-308
// Listing 13-11: Logging a subset of log entries to limit CPU and I/O overhead
func Example_zapSampling() {
	zl := zap.New(
		// The NewSamplerWithOptions function wraps zap.Core with sampling
		// functionality.
		zapcore.NewSamplerWithOptions(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderCfg),
				zapcore.Lock(os.Stdout),
				zapcore.DebugLevel,
			),
			// It requires three additional arguments:
			// - A sampling interval
			// - The number of initial duplicate log entries to record
			// - An integer representing the nth duplicate log entry to record after
			//   that point.
			// In this example, you are logging the first log entry, and then
			// every third duplicate for the remainder of the one-second interval.
			time.Second, 1, 3,
		),
	)
	defer func() { _ = zl.Sync() }()

	// You make 10 iterations around a loop.
	for i := 0; i < 10; i++ {
		if i == 5 {
			// On the sixth iteration, the example sleeps for one second to
			// ensure that the sample logger starts logging anew during the next
			// one second interval.
			time.Sleep(time.Second)
		}
		// Each iteration logs both the counter and a generic debug message,
		// which stays the same for each iteration.
		zl.Debug(fmt.Sprintf("%d", i))
		zl.Debug("debug message")
	}

	// Examining the output, you see that the debug message prints during the
	// first iteration and not again until the log encounters the third
	// duplicate debug message during the fourth loop iteration.
	// But on the sixth iteration, the example sleeps, and the sample logger
	// ticks over to the next one-second interval, starting the logging over.
	// It logs the first debug message of the interval in the sixth loop
	// iteration and the third duplicate debug message in the ninth iteration of
	// the loop.

	// Output:
	// {"level":"debug","msg":"0"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"1"}
	// {"level":"debug","msg":"2"}
	// {"level":"debug","msg":"3"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"4"}
	// {"level":"debug","msg":"5"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"6"}
	// {"level":"debug","msg":"7"}
	// {"level":"debug","msg":"8"}
	// {"level":"debug","msg":"debug message"}
	// {"level":"debug","msg":"9"}
}

// Pages 309-310
// Listing 13-12: Creating the new logger using an atomic leveler.
func Example_zapDynamicDebugging() {
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tempDir) }()

	// Your code will watch for the level.debug file in the temporary directory.
	// When the file is present, you'll dynamically change the logger's level to debug.
	debugLevelFile := filepath.Join(tempDir, "level.debug")
	// To do that, you need a new atomic leveler.
	// By default, the atomic leveler uses the info level, which suites this
	// example just fine.
	atomicLevel := zap.NewAtomicLevel()

	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			// You pass in the atomic leveler when creating the core as opposed
			// to specifying a log file itself.
			atomicLevel,
		),
	)
	defer func() { _ = zl.Sync() }()

	// Pages 310-311
	// Listing 13-13: Watching for any changes to the semaphore file.
	// First, you create a filesystem watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = watcher.Close() }()

	// You'll use the watcher to watch the temporary directory.
	// The watcher will notify you of any changes to or within that directory.
	err = watcher.Add(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	ready := make(chan struct{})
	go func() {
		defer close(ready)

		// You'll also want to capture the log level so that you can revert to
		// it when you remove the semaphore file.
		originalLevel := atomicLevel.Level()

		for {
			select {
			// Next, you listen for events from the watcher.
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Since you're watching a directory, you filter out any event
				// unrelated to the semaphore file itself.
				// Even then, you're only interested in the creation or removal
				// of the semaphore file.
				if event.Name == debugLevelFile {
					switch {
					// If you receive a semaphore file, you change the atomic
					// leveler's level to debug.
					case event.Op&fsnotify.Create == fsnotify.Create:
						atomicLevel.SetLevel(zapcore.DebugLevel)
						ready <- struct{}{}
					// If you receive a semaphore file removal event, then you
					// set the atomic leveler's level back to its original level.
					case event.Op&fsnotify.Remove == fsnotify.Remove:
						atomicLevel.SetLevel(originalLevel)
						ready <- struct{}{}
					}
				}
			// If you receive an error from the watcher at any point, you log it
			// at the error level.
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				zl.Error(err.Error())
			}
		}
	}()

	// Pages 311-312
	// Listing 13-14: Testing the logger's use of the semaphore file.
	// The logger's current log level via the atomic leveler is info.
	// Therefore, the logger does not write the initial debug log entry to
	// standard output.
	zl.Debug("this is below the logger's threshold")

	// But if you create the semaphore file, the code in Listing 13-13 should
	// dynamically change the logger's level to debug.
	df, err := os.Create(debugLevelFile)
	if err != nil {
		log.Fatal(err)
	}
	err = df.Close()
	if err != nil {
		log.Fatal(err)
	}
	<-ready

	// If you add another debug log entry, the logger should write it to the
	// standard output.
	zl.Debug("this is now at the logger's threshold")

	// You then remove the semaphore file.
	err = os.Remove(debugLevelFile)
	if err != nil {
		log.Fatal(err)
	}
	<-ready

	// And then write both a debug log entry and an info log entry.
	// Since the semaphore file no longer exists, the logger should write only
	// the info log entry to standard output.
	zl.Debug("this is below the logger's threshold again")
	zl.Info("this is at the logger's current threshold")

	// Output:
	// {"level":"debug","msg":"this is now at the logger's threshold"}
	// {"level":"info","msg":"this is at the logger's current threshold"}
}
