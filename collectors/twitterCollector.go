package collectors

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	log "github.com/sirupsen/logrus"
	"time"
)

type TwitterCollector struct{}

func (t TwitterCollector) sourceName() string {
	return "twitter"
}

func (t TwitterCollector) collectEvent() (string, string) {
	config := oauth1.NewConfig("qmHP2muP1cshDiYk1hHOTP1tN", "51XgOonYmwlPeqfkTHd6OA89AihLJ8Y5t6M684U64Vo3g82OfX")
	token := oauth1.NewToken("937756850174545920-q0oGAyeCZ8wHrKSBFLVTgpOhJ1b8AAY", "oo0Gk6VPSyZ7N3eyzNy0adO7p4ABCv6ze2XuChRWtJHRF")
	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	var tweets []string

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		if tweet.RetweetedStatus == nil {
			tweets = append(tweets, tweet.Text)
		}
	}

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Track:         []string{"mall"},
		StallWarnings: twitter.Bool(true),
		Locations:     []string{"-76.8507235", "-55.1671700", "-66.6756380", "-17.5227345"},
	}
	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Error("Failed to get Twitter event")
		return "0", "0"
	}

	go demux.HandleChan(stream.Messages)

	time.Sleep(15 * time.Second)
	stream.Stop()

	var allTweets string
	for idx, l := range tweets {
		if idx == 0 {
			allTweets = allTweets + l
		} else {
			allTweets = allTweets + "#" + l
		}
	}

	return allTweets, "0"
}

func (t TwitterCollector) estimateEntropy() int {
	return 0
}

func (t TwitterCollector) sourceID() int {
	return 2
}
