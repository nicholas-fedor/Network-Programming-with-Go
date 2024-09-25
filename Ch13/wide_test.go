// Pages 312-
// Listing 13-15: Creating a ResponseWriter to capture the response status code
// and length.
package Ch13

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type wideResponseWriter struct {
	// The new type embeds an object that implements the http.ResponseWriter
	// interface.
	http.ResponseWriter
	// In addition, you add length and status fields, since those values are
	// ultimately what you want to log from the response.
	length, status int
}

// You override the WriteHeader method to easily capture the status code.
func (w *wideResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

// Likewise, you override the Write method to keep an accurate accounting of the
// number of written bytes.
func (w *wideResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += n

	if w.status == 0 {
		// You optionally set the status code should the caller execute Write
		// before WriteHeader
		w.status = http.StatusOK
	}

	return n, err
}

// Page 313
// Listing 13-16: Implementing wide event logging middleware.
// The wide event logging middleware accepts both a *zap.Logger and
// an http.Handler and returns an http.Handler.
func WideEventLog(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// First, you embed the http.ResponseWriter in a new instance of
			// your wide event logging-aware response writer.
			wideWriter := &wideResponseWriter{ResponseWriter: w}

			// Then, you call the ServeHTTP method of the next http.Handler,
			// passing in your response writer.
			next.ServeHTTP(wideWriter, r)

			addr, _, _ := net.SplitHostPort(r.RemoteAddr)
			// Finally, you make a single log entry with the various bits of
			// data about the request and response. Note that we're taking care
			// to omit values that would change with each execution and break
			// the example output, such as call duration. You would likely have
			// to write code to deal with these in a real implementation.
			logger.Info("example wide event",
				zap.Int("status code", wideWriter.status),
				zap.Int("response length", wideWriter.length),
				zap.Int64("content_length", r.ContentLength),
				zap.String("method", r.Method),
				zap.String("proto", r.Proto),
				zap.String("remote_addr", addr),
				zap.String("uri", r.RequestURI),
				zap.String("user_agent", r.UserAgent()),
			)
		},
	)
}

// Page 314
// Listing 13-17: Using the wide event logging middleware ot log the details of
// a GET call.
func Example_wideLogEntry() {
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zap.DebugLevel,
		),
	)
	defer func() { _ = zl.Sync() }()

	ts := httptest.NewServer(
		// You pass *zap.Logger into the middleware as the first argument and
		// http.Handler as the second argument.
		WideEventLog(zl, http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer func(r io.ReadCloser) {
					_, _ = io.Copy(io.Discard, r)
					_ = r.Close()
				}(r.Body)
				// The handler writes a simple Hello! to the response so the
				// response length is nonzero.
				// That way, you can prove that your response writer works.
				_, _ = w.Write([]byte("Hello!"))
			},
		)),
	)
	defer ts.Close()

	// The logger writes the log entry immediately before you receive the
	// response to your GET request.
	resp, err := http.Get(ts.URL + "test")
	if err != nil {
		// Since this is just an example, the logger's Fatal method is used,
		// which writes the error message to the log file and calls os.Exit(1)
		// to terminate the application.
		// This shouldn't be used in code that is supposed to keep running in the
		// event of an error.
		zl.Fatal(err.Error())
	}
	_ = resp.Body.Close()

	// Output:
	// {"level":"info","msg":"example wide event","status_code":200,"response_length":6,"content_length":0,"method":"GET","proto":"HTTP/1.1","remote_addr":"127.0.0.1","uri":"/test","user_agent":"Go-http-client/1.1"}
}

// Pages 315-316
// Listing 13-18:
func TestZapLogRotation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			// Like the *bytes.Buffer in listing 13-9, *lumberjack.Logger
			// does not implement the zapcore.WriteSyncer.
			// It too lacks a sync method.
			// Therefore, you need to wrap it in a call to zapcore.AddSync.
			zapcore.AddSync(
				// Lumberjack includes several field to configure its behavior,
				// though its defaults are sensible.
				&lumberjack.Logger{
					// It uses a log filename in the format
					// <processname>-lumberjack.log, saved in the temporary
					// directory, unless you explicitly give it a log filename.
					Filename:   filepath.Join(tempDir, "debug.log"),
					// You can elect to save hard drive space and have
					// Lumberjack compress rotated log files.
					Compress:   true,
					// Each rotated log file is timestamped using UTC by
					// default, but you can instruct Lumberjack to use local
					// time instead.
					LocalTime:  true,
					// You can configure the maximum log file age before
					// it should be rotated.
					MaxAge:     7,
					// The maximum number of rotated log files to keep.
					MaxBackups: 5,
					// The maximum size in megabytes of a log file before it
					// should be rotated.
					MaxSize:    100,
				},
			),
			zapcore.DebugLevel,
		),
	)
	defer func() { _ = zl.Sync() }()

	zl.Debug("debug message written to the log file")
}
