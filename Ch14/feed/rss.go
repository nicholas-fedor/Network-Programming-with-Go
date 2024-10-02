// Pages 330-331
// Listing 14-1: Structure that represents the XKCD RSS Feed.

package feed

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// Go's encoding/xml package can use struct tags to map XML tags to their
// corresponding struct fields.

// The Item struct represents each item (comic) in the feed.
type Item struct {
	Title     string `xml:"title"`
	URL       string `xml:"link"`
	Published string `xml:"pubDate"`
}

// The RSS struct represents the RSS feed.
type RSS struct {
	Channel struct {
		Items []Item `xml:"item"`
	} `xml:"channel"`
	// It's important to keep track of the feed's entity tag.
	// Web servers often derive entity tags for content that may not change from
	// one request to another.
	// Clients can keep track of these entity tags and present them with future
	// requests.
	// If the server determines that the requested content has the same entity
	// tag, it can forego returning the entire payload and return a 304 Not
	// Modified status code so the client knows to use its cached copy instead.
	entityTag string
}

// Pages 331-332
// Listing 14-2: Methods to parse the XKCD RSS feed and return a slice of items.
// There are three things to note here:
// 1) The RSS struct and its methods are not safe for concurrent use.
// 2) The Items method returns a slice of items in the RSS struct, which is
// empty until the code calls the ParseURL method to populate the RSS struct.
// 3) The the Items method makes a copy of the Items slice and returns the copy
// to prevent possible corruption of the original Items slice. This is a bit
// overkill for this use case, but it's best to be aware that you're returning a
// reference type that the receiver can modify. If the receiver modifies the
// copy, it won't affect your original.

func (r RSS) Items() []Item {
	items := make([]Item, len(r.Channel.Items))
	copy(items, r.Channel.Items)

	return items
}

// The ParseURL method retrieves the RSS feed by using a GET call.
// If the feed is new, the method reads the XML from the response body and
// invokes the xml.Unmarshal function to populate the RSS struct with the XML in
// the server.
func (r *RSS) ParseURL(ctx context.Context, u string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	// Notice the request's ETag header is conditionally set, so the XKCD server
	// can determine whether it needs to send the feed contents or you currently
	// have the latest version.
	if r.entityTag != "" {
		req.Header.Add("ETag", r.entityTag)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	// If the server responds with a 304 Not Modified status code, then the RSS
	// struct remains unchanged.
	case http.StatusNotModified: // no-op
	// If the response status code is 200 OK, then you received a new version of
	// the feed and unmarshal the response body's XML into the RSS struct.
	case http.StatusOK:
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()

		err = xml.Unmarshal(b, r)
		if err != nil {
			return err
		}

		// If successful, you update the entity tag.
		r.entityTag = resp.Header.Get("ETag")
	default:
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	// With this logic in place, the RSS struct should update itself only if its
	// entity tag is empty, as it would be on initialization of the struct, or
	// if a new feed is available.

	return nil
}