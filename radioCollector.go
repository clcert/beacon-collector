package main

import (
	"bufio"
	"encoding/hex"
	"net/http"
)

type RadioCollector struct{}

func (r RadioCollector) collectEvent() string {
	streamURL := "http://200.89.71.21:8000/;"
	resp, _ := http.Get(streamURL)

	reader := bufio.NewReader(resp.Body)
	var audioBytes []byte
	for i := 0; i <= 99839; i++ {
		b, _ := reader.ReadByte()
		audioBytes = append(audioBytes, b)
	}
	//fileMP3, _ := os.Create("audio.mp3")
	//defer fileMP3.Close()
	//fileMP3.Write(audioBytes)

	audioHex := hex.EncodeToString(audioBytes)
	return audioHex
}

func (r RadioCollector) estimateEntropy() int {
	return 0
}

func (r RadioCollector) sourceID() int {
	return 3
}
