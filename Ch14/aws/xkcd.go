// Page 337
// Listing 14-4: Creating persistent variables and request and response types
package main

import (
	"context"

	"Ch14/feed"
	"github.com/aws/aws-lambda-go/lambda"
)

// We are specifying variables at the package level that will persist between
// function calls while the function persists in memory.

// We define a feed object and the URL of the RSS feed.
// Populating a new feed.RSS object involves a bit of overhead.
// We can avoid that overhead on subsequent function calls if we store the
// object in a variable at the package level so it lives beyond each function
// call.
// This allows us to take advantage of the entity tag support in feed.RSS.
var (
	rssFeed feed.RSS
	feedURL = "https://xkcd.com/rss.xml"
)

// The EventRequest and EventResponse types define the format of a client
// request and the function's response.
// ASW Lambda unmarshals the JSON from the client's HTTP request body into the
// EventRequest object and marshals the the function's EventResponse to JSON to
// the HTTP response body before returning it to the client.
type EventRequest struct {
	Previous bool `json:"previous"`
}

type EventResponse struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Published string `json:"published"`
}

// Page 338
// Listing 14-5: Main function and first part of the Lambda function named LatestXKCD
func main() {
	// Hook the function into Lambda by passing it to the lambda.Start method.
	// Instantiate dependencies in an init function, or before this statement,
	// if the function requires it.
	lambda.Start(LatestXKCD)
}

// The LatestXKCD function accepts a context and an EventRequest and returns an
// EventResponse and an error interface.
func LatestXKCD(ctx context.Context, req EventRequest) (EventResponse, error) {
	// It defines a response object with default Title and URL values.
	// The function returns the response as is in the event of an error or an
	// empty feed.
	resp := EventResponse{Title: "xkcd.com", URL: "https://xkcd.com/"}

	// Parsing the feed URL populates the rssFeed object with the latest feed details.
	if err := rssFeed.ParseURL(ctx, feedURL); err != nil {
		return resp, err
	}

	// Page 338
	// Listing 14-6: Populating the response with the feed results.

	// If the client requests the previous XKCD comic and there are at least two
	// feed items, the function populates the response with details of the
	// previous XKCD comic.
	// Otherwise, the function populates the response with the most recent XKCD
	// comic details, provided there's at least one feed item.
	// If neither case is true, the client receives the response with its
	// default values from Listing 14-5.
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

	return resp, nil
}
