package collectors

import (
	"bufio"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
	"net/http"
)

type RadioCollector struct{}

func (r RadioCollector) collectEvent() (string, string) {
	streamURL := "http://200.89.71.21:8000/;"
	resp, _ := http.Get(streamURL)

	if resp == nil {
		log.Error("Failed to get Radio event")
		return "0", "0"
	}
	reader := bufio.NewReader(resp.Body)
	var firstFrame []byte
	var audioBytes []byte
	var frameSize = 365
	var padding = false
	for i := 0; i <= 99839; i++ {
		b, _ := reader.ReadByte()
		if i == 2 {
			if b == 130 {
				padding = true
			}
		}
		if i < frameSize {
			firstFrame = append(firstFrame, b)
		}
		if i == frameSize {
			if padding {
				firstFrame = append(firstFrame, b)
			}
		}
		audioBytes = append(audioBytes, b)
	}
	//fileMP3, _ := os.Create("audio.mp3")
	//defer fileMP3.Close()
	//fileMP3.Write(audioBytes)
	firstFrameHashedHex := fmt.Sprintf("%x", sha3.Sum512(firstFrame))
	audioHex := hex.EncodeToString(audioBytes)
	return audioHex, firstFrameHashedHex
}

func (r RadioCollector) estimateEntropy() int {
	return 0
}

func (r RadioCollector) sourceID() int {
	return 3
}
