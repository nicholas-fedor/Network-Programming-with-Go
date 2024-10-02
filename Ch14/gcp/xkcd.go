// Pages 342-343
// Listing 14-7: Creating persistent variables and request and response types.
package gcp

import (
	"encoding/json"
	"log"
	"net/http"

	"Ch14/feed"
)

var (
	rssFeed feed.RSS
	feedURL = "https://xkcd.com/rss.xml"
)

type EventRequest struct {
	Previous bool `json:"previous"`
}

type EventResponse struct {
	Title     string `json:"response"`
	URL       string `json:"url"`
	Published string `json:"published"`
}

// The types are identical to those for AWS Lambda; however, GCP Functions won't
// unmarshal the request body into an EventRequest for you.

// Page 343
// Listing 14-8: Handling the request and response and optionally updating the
// RSS feed.
// The LatestXKCD function refreshes the RSS feed by using the ParseURL method.
func LatestXKCD(w http.ResponseWriter, r *http.Request) {
	var req EventRequest
	resp := EventResponse{Title: "xkcd.com", URL: "https://xkcd.com/"}

	// Unlike the AWS code, we need to JSON-unmarshal the request body and
	// marshal the response to JSON before sending it to the client.
	defer func() {
		w.Header().Set("Content-Type", "application/json")
		out, _ := json.Marshal(&resp)
		_, _ = w.Write(out)
	}()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("decoding request: %v", err)
		return
	}

	// Even though LatestXKCD doesn't receive a context in its function
	// parameters, we can use the request's context to cancel the parser if the
	// socket connection with the client terminates before the parser returns.
	if err := rssFeed.ParseURL(r.Context(), feedURL); err != nil {
		log.Printf("parsing feed: %v:", err)
		return
	}

	// Page 344
	// Listing 14-9: Populating the response with the feed results.
	switch items := rssFeed.Items(); {
	case req.Previous && len(items) > 1:
		resp.Title = items[1].Title
		resp.URL = items[1].URL
		resp.Published = items[1].Published
	case len(items) > 0:
		resp.Title = items[0].Title
		resp.URL = items[0].URL
		resp.Published = items[0].Published
	}
}