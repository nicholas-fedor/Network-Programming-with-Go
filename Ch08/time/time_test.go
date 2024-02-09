// Page 174
// Listing 8-2: Retrieving a timestamp from time.gov
package main

import (
	"net/http"
	"testing"
	"time"
)

func TestHeadTime(t *testing.T) {
	// The net/http package includes a few helper function to make GET, HEAD,
	// or POST requests. Here, we use the http.Get function to
	// https://www.time.gov/ to retrieve the default resource.
	// Go's HTTP client automatically upgrades to HTTPS for you because that's
	// the protocol indicated by the URL's scheme.
	resp, err := http.Head("https://www.time.gov/")
	if err != nil {
		t.Fatal(err)
	}
	// Although you don't read the contents of the response body, you must close it.
	_ = resp.Body.Close() // Always close this without exception.

	now := time.Now()
	// Now that you have a response, you retrieve the Date header, which
	// indicates the time at which the server created the response.
	// You can then use this value to calculate the clock skew of your computer.
	// Granted, you lose accuracy because of latency between the server's
	// generating the header and your code's processing of it, as well as the
	// lack of nanosecond resolution of the Date header itself.
	date := resp.Header.Get("Date")
	if date == "" {
		t.Fatal("no Date header received from time.gov")
	}

	dt, err := time.Parse(time.RFC1123, date)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("time.gov: %s (skew %s)", dt, now.Sub(dt))
}
