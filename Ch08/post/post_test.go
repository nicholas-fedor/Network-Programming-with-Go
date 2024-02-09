// Pages 179-180
// Listing 8-6: A handler that can decode JSON into a User object.
// Listing 8-6 creates a new type named User that will encode to JSON and post
// to the handler.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type User struct {
	First string
	Last  string
}

// The handlePostUser function returns a function that will handle POST requests.
func handlePostUser(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(r io.ReadCloser) {
			// Unlike the Go HTTP client, the Go HTTP server must explicitly
			// drain the request body before closing it.
			_, _ = io.Copy(io.Discard, r)
			_ = r.Close()
		}(r.Body)

		// If the request method is anything other than POST.
		if r.Method != http.MethodPost {
			// It returns a status code indicating that the server disallows the method.
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		// The function then attempts to decode the JSON in the request body to
		// a User object.
		var u User
		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			t.Error(err)
			http.Error(w, "Decode Failed", http.StatusBadRequest)
			return
		}

		// If successful, the response's status is set to Accepted.
		w.WriteHeader(http.StatusAccepted)
	}
}

// Pages 180-181
// Listing 8-7: Encoding a User object to JSON and POST to the test server.
// The test encodes a User object into JSON and sends it in a POST request to
// the test server.
func TestPostUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handlePostUser(t)))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	// The test first makes sure that the test server's handler properly
	// responds with an error if the client sends the wrong type of request.
	// If the test server receives anything other than a POST request, it will
	// respond with a Method Not Allowed error.
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d; actual status %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	u := User{First: "Adam", Last: "Woodbeck"}
	// Then, the test encodes a User object into JSON and writes the data to a
	// bytes buffer.
	err = json.NewEncoder(buf).Encode(&u)
	if err != nil {
		t.Fatal(err)
	}

	// It makes a POST request to the test server's URL with the content type
	// application/json because the bytes buffer, representing the request body,
	// contains JSON.
	// The content type informs the server's handler about the type of data to
	// expect in the request body.
	resp, err = http.Post(ts.URL, "application/json", buf)
	if err != nil {
		t.Fatal(err)
	}

	// If the server's handler properly decoded the request body, the response
	// status code is 202 Accepted.
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status %d; actual status %d", http.StatusAccepted, resp.StatusCode)
	}

	_ = resp.Body.Close()
}

// Page 181
// Listing 8-8: Creating a new request body, multipart writer, and write form
// data.
// Listing 8-8 introduces a new test that walks you through the process of
// building up a multipart request body using the mime/multipart package.
func TestMultipartPost(t *testing.T) {
	// First, you create a new buffer to act as the request body.
	reqBody := new(bytes.Buffer)
	// You then create a new multipart writer that wraps the buffer.
	// The multipart writer generates a random boundary upon initialization.
	w := multipart.NewWriter(reqBody)

	for k, v := range map[string]string{
		"date":        time.Now().Format(time.RFC3339),
		"description": "Form values with attached files",
	} {
		// Finally, you write form fields into its own part, writing the
		// boundary, appropriate headers, and the form field value to each
		// part's body.
		// At this point, your request body has two parts, one for the date form
		// field and one for the description form field.
		err := w.WriteField(k, v)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Page 182
	// Listing 8-9: Writing two files to the request body, each in its own MIME
	// part.
	// Attaching a field to a request body isn't as straightforward as adding
	// form field data.

	for i, file := range []string{
		"./files/hello.txt",
		"./files/goodbye.txt",
	} {
		// First, you need to create a multipart section writer from Listing
		// 8-8's multipart writer.
		// The CreateFormField method accepts a field name and a filename.
		// The server uses this filename when parsing the MIME part.
		// It does not need to match the filename you attach.
		filePart, err := w.CreateFormFile(fmt.Sprintf("file%d", i+1), filepath.Base(file))
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}

		// Now, you just open the file and copy its contents to the MIME part writer.
		_, err = io.Copy(filePart, f)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	// When you're done adding parts to the request body, you must close the
	// multipart writer, which finalizes the request body by appending the boundary.
	err := w.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Page 183
	// Listing 8-10: Sending a POST request to httpbin.org with Go's default
	// HTTP client.
	// Listing 8-10 posts the request to a well-regarded test server,
	// httpbin.org.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// First, you create a new request and pass it a context that will time out
	// in 60 seconds.
	// Since you're making this call over the internet, you don't have as much
	// certainty that your request will reach its destination as you do when
	// testing over localhost.
	// The POST request is destined for https://www.httpbin.org/ and will send
	// the multipart request body in its payload.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://httpbin.org/post", reqBody)
	if err != nil {
		t.Fatal(err)
	}
	// Before you send the request, you need to set the Content-Type header to
	// inform the web server you're sending multiple parts in this request.
	// The multipart writer's FormDataContentType method returns the appropriate
	// Content-Type value that includes its boundary.
	// The web server uses the boundary from this header to determine where one
	// part stops and another starts as it reads the request body.
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = req.Body.Close() }()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d; actual status %d", http.StatusOK, resp.StatusCode)
	}

	t.Logf("\n%s", b)
}
