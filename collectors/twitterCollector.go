package collectors

import (
	"bufio"
	"bytes"
	"container/heap"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type TwitterCollector struct{}

func (t TwitterCollector) sourceName() string {
	return "twitter"
}

type CollectedTweet struct {
	Id        string `json:"id"`
	CreatedAt string `json:"created_at"`
	AuthorId  string `json:"author_id"`
	Text      string `json:"text"`
}

func (c CollectedTweet) String() string {
	return fmt.Sprintf("[%s;%s;%s;%s]", c.Id, c.CreatedAt, c.AuthorId, c.Text)
}

type TweetsHeap []CollectedTweet

func (t TweetsHeap) Len() int {
	return len(t)
}

func (t TweetsHeap) Less(i, j int) bool {
	firstTweet := t[i]
	secondTweet := t[j]
	firstDate, _ := time.Parse(time.RFC3339, firstTweet.CreatedAt)
	secondDate, _ := time.Parse(time.RFC3339, secondTweet.CreatedAt)

	if firstDate.Before(secondDate) {
		return true
	}
	if firstTweet.Id < secondTweet.Id {
		return true
	}
	return false
}

func (t TweetsHeap) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t *TweetsHeap) Push(x interface{}) {
	*t = append(*t, x.(CollectedTweet))
}

func (t *TweetsHeap) Pop() interface{} {
	old := *t
	n := len(old)
	x := old[n-1]
	*t = old[0 : n-1]
	return x
}

func getTwitterCredentials() map[string]string {
	configJsonFile, err := os.Open("collectors/twitterConfig.json")
	if err != nil {
		fmt.Println(err)
	}
	defer configJsonFile.Close()
	twitterCredentials := make(map[string]string)
	byteValue, _ := ioutil.ReadAll(configJsonFile)
	json.Unmarshal(byteValue, &twitterCredentials)

	return twitterCredentials
}

func (t TwitterCollector) collectEvent(ch chan Event) {
	// currentMinute := time.Now().UTC().Minute()
	currentTimestamp := time.Now().UTC()
	const startSecondMark = 5
	const extractingDuration = 10
	const waitingAfter = 3
	startLimitTimestamp := time.Date(currentTimestamp.Year(), currentTimestamp.Month(), currentTimestamp.Day(), currentTimestamp.Hour(), currentTimestamp.Minute(), startSecondMark, 0, currentTimestamp.Location())
	endLimitTimestamp := startLimitTimestamp.Add(extractingDuration * time.Second)

	twitterCredentials := getTwitterCredentials()
	var consumerKey = twitterCredentials["consumer_key"]
	var consumerSecret = twitterCredentials["consumer_secret"]
	bearerToken := getBearerToken(consumerKey, consumerSecret)

	var streamURL = "https://api.twitter.com/2/tweets/sample/stream?tweet.fields=created_at&expansions=author_id"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", streamURL, nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Error("twitter response error, status code: " + strconv.Itoa(resp.StatusCode))
		// return "", "", 2
		ch <- Event{"", "", 2}
		return
	}

	tweetReader := bufio.NewReader(resp.Body)

	tweets := &TweetsHeap{}
	s := bufio.NewScanner(tweetReader)
	heap.Init(tweets)
	//aux := false
	for s.Scan() {
		collectedTweet := map[string]CollectedTweet{"data": {}}
		_ = json.Unmarshal(s.Bytes(), &collectedTweet)
		collectedTweetCreatedAt, _ := time.Parse(time.RFC3339, collectedTweet["data"].CreatedAt)
		if collectedTweetCreatedAt.Equal(startLimitTimestamp) || (collectedTweetCreatedAt.After(startLimitTimestamp) && collectedTweetCreatedAt.Before(endLimitTimestamp)) || collectedTweetCreatedAt.Equal(endLimitTimestamp) {
			heap.Push(tweets, collectedTweet["data"])
		}
		//
		//if collectedTweetCreatedAt.After(endLimitTimestamp) {
		//	aux = true
		//}
		//if aux {
		//	if collectedTweetCreatedAt.Before(endLimitTimestamp) {
		//		fmt.Println("arrived late tweet!")
		//	}
		//}
		//
		if collectedTweetCreatedAt.Equal(endLimitTimestamp.Add(waitingAfter * time.Second)) {
			break
		}
		//if currentMinute == collectedTweetCreatedAt.Minute() && startSecondMark <= collectedTweetCreatedAt.Second() && collectedTweetCreatedAt.Second() <= (startSecondMark+extractingDuration) {
		//	heap.Push(tweets, collectedTweet["data"])
		//}
		//if collectedTweetCreatedAt.Second() == (startSecondMark+extractingDuration)+5 {
		//	break
		//}
	}

	var tweetsResponse []CollectedTweet
	var firstTimestamp string
	for tweets.Len() > 0 {
		tweetsResponse = append(tweetsResponse, heap.Pop(tweets).(CollectedTweet))
		if firstTimestamp == "" {
			firstTimestamp = tweetsResponse[0].CreatedAt
		}
	}
	tweetsAsJSONBytes, _ := json.Marshal(tweetsResponse)
	tweetsAsJSONString := string(tweetsAsJSONBytes)

	// return tweetsAsJSONString, firstTimestamp, 0
	ch <- Event{tweetsAsJSONString, firstTimestamp, 0}
}

func twitterCanonicalForm(t []CollectedTweet) string {
	var response string
	for i := 0; i < len(t); i++ {
		tweet := t[i]
		response += tweet.CreatedAt + tweet.Id + tweet.AuthorId + tweet.Text
	}
	return response
}

func getBearerToken(consumerKey string, consumerSecret string) string {
	credentials := []string{consumerKey, consumerSecret}
	credentialsBase64 := base64.StdEncoding.EncodeToString([]byte(strings.Join(credentials, ":")))
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://api.twitter.com/oauth2/token", bytes.NewBuffer([]byte("grant_type=client_credentials")))
	req.Header.Add("Authorization", "Basic "+credentialsBase64)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)
	authInfo := make(map[string]string)
	_ = json.Unmarshal(response, &authInfo)

	return authInfo["access_token"]
}

func (t TwitterCollector) estimateEntropy() int {
	return 0
}

func (t TwitterCollector) getCanonicalForm(s string) string {
	if s == "" {
		return ""
	}
	var tweets []CollectedTweet
	err := json.Unmarshal([]byte(s), &tweets)
	if err != nil {
		log.Error(err)
	}
	var response = twitterCanonicalForm(tweets)
	return response
}
