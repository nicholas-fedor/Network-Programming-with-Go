// Page 194
// Listing 9-4: The default handler implementation.
// The handlers.DefaultHandler function returns a function converted to the
// http.HandlerFunc type.
// The http.HandlerFunc type implements the http.Handler interface.
// Go programmers commonly convert a function with the signature func(w
// http.ResponseWriter, r *http.Request) to the http.HandlerFunc type so the
// function implements the http.Handler interface.
package handlers

import (
	"html/template"
	"io"
	"net/http"
)

// This code could have a security vulnerability since part of the response body
// might come from the request body.
// A malicious client can send a request payload that includes JavaScript, which
// could run on a client's computer.
// This behavior can lead to an XSS attack.
// To prevent these attacks, you must properly escape all client-supplied
// content before sending it in a response.
// Here, you use the html/template package to create a simple template that
// reads "Hello, {{.}}!", where "{{.}}" is a placeholder for part of your
// response.
var t = template.Must(template.New("hello").Parse("Hello, {{.}}!"))

func DefaultHandler() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// The first bit of code you see is a deferred function that drains
			// and closes the request body.
			// Just as it's important for the client to drain and close the
			// response body to reuse the TCP session, it's important for the
			// server to do the same with the request body.
			// But unlike the Go HTTP client, closing the request body does not
			// implicitly drain it.
			// Granted, the http.Server will close the request body for you, but
			// it won't drain it.
			// To make sure you can reuse the TCP session, I recommend you drain
			// the request body at a minimum.
			// Closing it is optional.
			defer func(r io.ReadCloser) {
				_, _ = io.Copy(io.Discard, r)
				_ = r.Close()
			}(r.Body)

			var b []byte

			// The handler response differently depending on the request method.
			// If the client sent a GET request, the handler writes "Hello,
			// friend!" to the response writer.
			// If the request method is a POST, the handler first reads the
			// entire request body.
			switch r.Method {
			case http.MethodGet:
				b = []byte("friend")
			case http.MethodPost:
				var err error
				b, err = io.ReadAll(r.Body)
				if err != nil {
					// If an error occurs while reading the request body, the handler
					// uses the http.Error function to succinctly write the message
					// "Internal server error" to the response body and set the response
					// status code to 500.
					// Otherwise, the handler returns a greeting using the request body
					// contents.
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			default:
				// If the handler receives any other request method, it responds
				// with a 405 "Method Not Allowed" status.
				// The 405 response is technically not RFC-compliant without an
				// Allow header showing which methods the handler accepts.
				// Finally, the handler writes the response body.
				// not RFC-compliant due to lack of "Allow" header
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Templates derived from the html/template package automatically escape HTML
			// characters when you populate them and write the results to the
			// response writer.
			// HTML-escaping explains the funky characters in Listing 9-2's
			// second test case.
			// The client's browser will properly display the characters instead
			// of interpreting them as part of the HTML and JavaScript in the
			// response body.
			// The bottom line is to always use the html/template package when
			// writing untrusted data to a response writer.
			_ = t.Execute(w, string(b))
		},
	)
}
