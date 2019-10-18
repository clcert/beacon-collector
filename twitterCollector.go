package main

import (
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"log"
)

type TwitterCollector struct{}

func (t TwitterCollector) collectEvent() string {
	fmt.Println("Twitter!")
	config := oauth1.NewConfig("qmHP2muP1cshDiYk1hHOTP1tN", "51XgOonYmwlPeqfkTHd6OA89AihLJ8Y5t6M684U64Vo3g82OfX")
	token := oauth1.NewToken("937756850174545920-V2BQWRx07NZ4g81hrAKXrctT9raolUo", "tsTM7E1rY3aqhlHOIlX692NfVGaLPVZPgJnJx7TrZ77hG")
	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		fmt.Println(tweet.Text)
	}

	fmt.Println("Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Track:         []string{"cat"},
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal(err)
	}

	go demux.HandleChan(stream.Messages)

	return ""
}

func (t TwitterCollector) estimateEntropy() int {
	return 0
}

func (t TwitterCollector) sourceID() int {
	return 2
}
